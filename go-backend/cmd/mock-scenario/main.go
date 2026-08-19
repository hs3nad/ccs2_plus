package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type roomState struct {
	RoomID           string `json:"room_id"`
	Floor            int    `json:"floor"`
	Status           string `json:"status"`
	StayMode         string `json:"stay_mode"`
	CaseCode         string `json:"case"`
	ControllerOnline bool   `json:"controller_online"`
	RemainingMinutes *int   `json:"remaining_minutes,omitempty"`
}

type auditItem struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	RoomID    string `json:"room_id"`
	Action    string `json:"action"`
	Actor     string `json:"actor"`
	Result    string `json:"result"`
	StayMode  string `json:"stay_mode,omitempty"`
	CaseCode  string `json:"case,omitempty"`
	Detail    string `json:"detail,omitempty"`
}

type alertItem struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	RoomID    string `json:"room_id"`
	Severity  string `json:"severity"`
	RuleCode  string `json:"rule_code"`
	Status    string `json:"status"`
	Title     string `json:"title"`
	Detail    string `json:"detail"`
}

type sseEvent struct {
	Event string
	Data  any
}

type server struct {
	scenario            string
	delayedRecoverAfter time.Duration

	mu     sync.RWMutex
	rooms  map[string]roomState
	audit  []auditItem
	alerts []alertItem

	subMu sync.RWMutex
	subs  map[chan sseEvent]struct{}
}

func newServer(scenario string, delayedRecoverAfter time.Duration) *server {
	remain := 120
	return &server{
		scenario:            scenario,
		delayedRecoverAfter: delayedRecoverAfter,
		rooms: map[string]roomState{
			"1205": {
				RoomID:           "1205",
				Floor:            12,
				Status:           "vacant",
				StayMode:         "",
				CaseCode:         "",
				ControllerOnline: true,
				RemainingMinutes: &remain,
			},
		},
		subs: make(map[chan sseEvent]struct{}),
	}
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", s.handlePing)
	mux.HandleFunc("/api/state", s.handleState)
	mux.HandleFunc("/api/audit/events", s.handleAudit)
	mux.HandleFunc("/api/alerts", s.handleAlerts)
	mux.HandleFunc("/api/rooms/", s.handleRooms)
	mux.HandleFunc("/events", s.handleEvents)
	return mux
}

func (s *server) handlePing(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":   "ok",
		"time":     time.Now().Format(time.RFC3339),
		"scenario": s.scenario,
	})
}

func (s *server) handleState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	rooms := make([]roomState, 0, len(s.rooms))
	summary := map[string]int{
		"vacant":   0,
		"occupied": 0,
		"cleaning": 0,
		"overstay": 0,
		"dnd":      0,
	}
	for _, room := range s.rooms {
		rooms = append(rooms, room)
		summary[room.Status]++
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"rooms":   rooms,
		"summary": summary,
	})
}

func (s *server) handleAudit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	roomID := r.URL.Query().Get("room_id")
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]auditItem, 0, len(s.audit))
	for _, item := range s.audit {
		if roomID != "" && item.RoomID != roomID {
			continue
		}
		items = append(items, item)
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *server) handleAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	items := append([]alertItem(nil), s.alerts...)
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *server) handleRooms(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/rooms/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 || r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	roomID := parts[0]
	op := parts[1]

	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	action := body["action"].(string)
	requestID := stringOrDefault(body["request_id"], fmt.Sprintf("req-%d", time.Now().UnixNano()))

	if err := s.applyAction(roomID, op, action, body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": map[string]any{
				"code":       "INVALID_ACTION",
				"message":    err.Error(),
				"request_id": requestID,
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"request_id": requestID,
		"room_id":    roomID,
		"action":     action,
		"status":     "accepted",
		"sent_at":    time.Now().Format(time.RFC3339),
	})

	if s.scenario == "sse-drop-after-full-state" {
		return
	}
	if s.scenario == "delayed-room-update" {
		s.publish(sseEvent{
			Event: "audit_event",
			Data:  s.latestAudit(),
		})
		if alert, ok := s.latestAlert(); ok {
			s.publish(sseEvent{
				Event: "fraud_alert",
				Data:  alert,
			})
		}
		return
	}
	if s.scenario == "delayed-room-update-recover" {
		s.publish(sseEvent{
			Event: "audit_event",
			Data:  s.latestAudit(),
		})
		if alert, ok := s.latestAlert(); ok {
			s.publish(sseEvent{
				Event: "fraud_alert",
				Data:  alert,
			})
		}

		go func(roomID, requestID, action string, delay time.Duration) {
			time.Sleep(delay)
			s.publish(sseEvent{
				Event: "room_update",
				Data:  s.roomPayload(roomID, requestID, action),
			})
		}(roomID, requestID, action, s.delayedRecoverAfter)
		return
	}

	s.publish(sseEvent{Event: "room_update", Data: s.roomPayload(roomID, requestID, action)})
	s.publish(sseEvent{Event: "audit_event", Data: s.latestAudit()})
	if alert, ok := s.latestAlert(); ok {
		s.publish(sseEvent{Event: "fraud_alert", Data: alert})
	}
}

func (s *server) applyAction(roomID, op, action string, body map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, ok := s.rooms[roomID]
	if !ok {
		return fmt.Errorf("room not found: %s", roomID)
	}

	switch op {
	case "checkin":
		room.Status = "occupied"
		room.StayMode = stringOrDefault(body["stay_mode"], "overnight")
		room.CaseCode = stringOrDefault(body["case"], room.StayMode)
	case "checkout":
		room.Status = "vacant"
		room.StayMode = ""
		room.CaseCode = ""
	case "extend":
		room.Status = "occupied"
		if stayMode := stringOrDefault(body["stay_mode"], ""); stayMode != "" {
			room.StayMode = stayMode
			room.CaseCode = stringOrDefault(body["case"], stayMode)
		}
	case "payment":
	case "command":
		switch action {
		case "start_cleaning":
			room.Status = "cleaning"
		case "finish_cleaning":
			room.Status = "vacant"
			room.StayMode = ""
			room.CaseCode = ""
		case "open_room", "service":
		default:
			return fmt.Errorf("unsupported action: %s", action)
		}
	default:
		return fmt.Errorf("unsupported route: %s", op)
	}

	s.rooms[roomID] = room

	now := time.Now().Format(time.RFC3339)
	s.audit = append([]auditItem{{
		ID:        stringOrDefault(body["request_id"], fmt.Sprintf("req-%d", time.Now().UnixNano())),
		Timestamp: now,
		RoomID:    roomID,
		Action:    action,
		Actor:     stringOrDefault(body["actor"], stringOrDefault(body["source"], "tester")),
		Result:    "accepted",
		StayMode:  room.StayMode,
		CaseCode:  room.CaseCode,
		Detail:    stringOrDefault(body["reason"], stringOrDefault(body["service_type"], "")),
	}}, s.audit...)

	if action == "open_room" && room.Status == "vacant" {
		s.alerts = append([]alertItem{{
			ID:        "alert-open-room-" + stringOrDefault(body["request_id"], "mock"),
			Timestamp: now,
			RoomID:    roomID,
			Severity:  "high",
			RuleCode:  "OPEN_ROOM_WHILE_VACANT",
			Status:    "open",
			Title:     "Open room while vacant",
			Detail:    "Open-room event was recorded while room state remained vacant.",
		}}, s.alerts...)
	}

	return nil
}

func (s *server) handleEvents(w http.ResponseWriter, r *http.Request) {
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

	ch := make(chan sseEvent, 16)
	s.subMu.Lock()
	s.subs[ch] = struct{}{}
	s.subMu.Unlock()
	defer func() {
		s.subMu.Lock()
		delete(s.subs, ch)
		close(ch)
		s.subMu.Unlock()
	}()

	_ = writeSSE(w, sseEvent{
		Event: "full_state",
		Data:  s.fullStatePayload(),
	})
	flusher.Flush()

	if s.scenario == "sse-drop-after-full-state" {
		return
	}

	for {
		select {
		case <-r.Context().Done():
			return
		case event := <-ch:
			if err := writeSSE(w, event); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func (s *server) publish(event sseEvent) {
	s.subMu.RLock()
	defer s.subMu.RUnlock()
	for ch := range s.subs {
		select {
		case ch <- event:
		default:
		}
	}
}

func (s *server) fullStatePayload() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rooms := make([]roomState, 0, len(s.rooms))
	summary := map[string]int{
		"vacant":   0,
		"occupied": 0,
		"cleaning": 0,
		"overstay": 0,
		"dnd":      0,
	}
	for _, room := range s.rooms {
		rooms = append(rooms, room)
		summary[room.Status]++
	}
	return map[string]any{
		"rooms":   rooms,
		"summary": summary,
	}
}

func (s *server) roomPayload(roomID, requestID, action string) map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	room := s.rooms[roomID]
	return map[string]any{
		"room_id":           room.RoomID,
		"floor":             room.Floor,
		"status":            room.Status,
		"stay_mode":         room.StayMode,
		"case":              room.CaseCode,
		"controller_online": room.ControllerOnline,
		"remaining_minutes": room.RemainingMinutes,
		"request_id":        requestID,
		"action":            action,
		"updated_at":        time.Now().Format(time.RFC3339),
	}
}

func (s *server) latestAudit() auditItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.audit) == 0 {
		return auditItem{}
	}
	return s.audit[0]
}

func (s *server) latestAlert() (alertItem, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.alerts) == 0 {
		return alertItem{}, false
	}
	return s.alerts[0], true
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
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

func stringOrDefault(v any, fallback string) string {
	if value, ok := v.(string); ok && value != "" {
		return value
	}
	return fallback
}

func main() {
	listenAddr := flag.String("listen", ":18080", "listen address")
	scenario := flag.String("scenario", "normal", "normal|delayed-room-update|delayed-room-update-recover|sse-drop-after-full-state")
	delayedRecoverAfter := flag.Duration("delayed-recover-after", 8*time.Second, "delay before sending room_update in delayed-room-update-recover")
	flag.Parse()

	srv := newServer(*scenario, *delayedRecoverAfter)
	log.Printf(
		"mock backend scenario listening on %s (scenario=%s delayed_recover_after=%s)",
		*listenAddr,
		*scenario,
		delayedRecoverAfter.String(),
	)
	log.Fatal(http.ListenAndServe(*listenAddr, srv.routes()))
}
