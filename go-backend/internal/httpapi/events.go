package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type sseEvent struct {
	Event string
	Data  any
}

type eventBroker struct {
	mu          sync.RWMutex
	subscribers map[chan sseEvent]struct{}
}

func newEventBroker() *eventBroker {
	return &eventBroker{
		subscribers: make(map[chan sseEvent]struct{}),
	}
}

func (b *eventBroker) Subscribe() (chan sseEvent, func()) {
	ch := make(chan sseEvent, 16)
	b.mu.Lock()
	b.subscribers[ch] = struct{}{}
	b.mu.Unlock()

	return ch, func() {
		b.mu.Lock()
		if _, ok := b.subscribers[ch]; ok {
			delete(b.subscribers, ch)
			close(ch)
		}
		b.mu.Unlock()
	}
}

func (b *eventBroker) Publish(event sseEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.subscribers {
		select {
		case ch <- event:
		default:
		}
	}
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch, unsubscribe := s.events.Subscribe()
	defer unsubscribe()

	if err := writeSSE(w, sseEvent{
		Event: "full_state",
		Data:  s.state.fullStatePayload(),
	}); err != nil {
		return
	}
	flusher.Flush()

	heartbeat := time.NewTicker(20 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			if err := writeSSE(w, event); err != nil {
				return
			}
			flusher.Flush()
		case <-heartbeat.C:
			if _, err := fmt.Fprint(w, ": keep-alive\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func writeSSE(w http.ResponseWriter, event sseEvent) error {
	payload, err := json.Marshal(event.Data)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "event: %s\n", event.Event); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", payload); err != nil {
		return err
	}
	return nil
}

func (s *Server) publishFullState() {
	s.events.Publish(sseEvent{
		Event: "full_state",
		Data:  s.state.fullStatePayload(),
	})
}

func (s *Server) publishRoomUpdate(roomID, requestID, action string) {
	s.events.Publish(sseEvent{
		Event: "room_update",
		Data:  s.state.roomUpdatePayload(roomID, requestID, action),
	})
}

func (s *Server) publishAuditEvent(entry auditEvent) {
	s.events.Publish(sseEvent{
		Event: "audit_event",
		Data:  entry,
	})
}

func (s *Server) publishFraudAlert(alert alertItem) {
	s.events.Publish(sseEvent{
		Event: "fraud_alert",
		Data:  alert,
	})
}

func (s *stateStore) fullStatePayload() map[string]any {
	items := s.list()
	summary := map[string]int{
		"vacant":   0,
		"occupied": 0,
		"cleaning": 0,
		"overstay": 0,
		"dnd":      0,
	}
	rooms := make([]roomSnapshot, 0, len(items))
	for _, item := range items {
		summary[item.Status]++
		rooms = append(rooms, item.snapshot())
	}
	return map[string]any{
		"rooms":   rooms,
		"summary": summary,
	}
}

func (s *stateStore) roomUpdatePayload(roomID, requestID, action string) map[string]any {
	room, ok := s.get(roomID)
	if !ok {
		return map[string]any{
			"room_id":    roomID,
			"request_id": requestID,
			"action":     action,
		}
	}
	payload := room.snapshot()
	return map[string]any{
		"room_id":           payload.RoomID,
		"floor":             payload.Floor,
		"status":            payload.Status,
		"stay_mode":         payload.StayMode,
		"case":              payload.CaseCode,
		"controller_online": payload.ControllerOnline,
		"remaining_minutes": payload.RemainingMinutes,
		"request_id":        requestID,
		"action":            action,
		"updated_at":        room.UpdatedAt.Format(time.RFC3339),
	}
}

func (s *stateStore) alertForAudit(entry auditEvent) (alertItem, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if entry.Action != "open_room" {
		return alertItem{}, false
	}

	room, ok := s.rooms[entry.RoomID]
	if !ok || room.Status != "vacant" {
		return alertItem{}, false
	}

	return alertItem{
		ID:        "alert-open-room-" + entry.ID,
		Timestamp: entry.Timestamp,
		RoomID:    entry.RoomID,
		Severity:  "high",
		RuleCode:  "OPEN_ROOM_WHILE_VACANT",
		Status:    "open",
		Title:     "Open room while vacant",
		Detail:    "Open-room event was recorded while room state remained vacant.",
	}, true
}
