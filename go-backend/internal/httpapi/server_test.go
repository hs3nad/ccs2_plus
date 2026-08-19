package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"ccs2plus/go-backend/internal/roommap"
	"ccs2plus/go-backend/internal/transport"
)

func newTestServer() (*Server, *transport.MemoryWriter) {
	repo := roommap.NewMemoryRepository([]roommap.RoomMapping{
		{
			RoomID:       "1205",
			ControllerID: "ctrl-1205",
			Address:      0x03,
			CardIndex:    3,
			PortIndex:    5,
		},
	})
	writer := &transport.MemoryWriter{}
	svc := transport.NewService(repo, writer)
	return NewServer(svc, repo), writer
}

func TestCheckInRouteBuildsLegacyFrames(t *testing.T) {
	server, writer := newTestServer()

	body, _ := json.Marshal(map[string]any{
		"request_id": "req-1",
		"action":     "check_in",
		"stay_mode":  "overnight",
		"case":       "overnight",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/rooms/1205/checkin", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if len(writer.Frames) != 2 {
		t.Fatalf("expected 2 frames, got %d", len(writer.Frames))
	}

	gotSlave := writer.Frames[0]
	wantSlave := []byte{0x3a, 0x03, 0x20, 0x00, 0xa3}
	if !bytes.Equal(gotSlave, wantSlave) {
		t.Fatalf("unexpected slave frame: got % X want % X", gotSlave, wantSlave)
	}

	gotDisplay := writer.Frames[1]
	wantDisplay := []byte{0x3a, 0xfe, 0x35, 0x03}
	if !bytes.Equal(gotDisplay, wantDisplay) {
		t.Fatalf("unexpected display frame: got % X want % X", gotDisplay, wantDisplay)
	}
}

func TestStartCleaningRouteBuildsLegacyFrames(t *testing.T) {
	server, writer := newTestServer()

	body, _ := json.Marshal(map[string]any{
		"request_id": "req-2",
		"action":     "start_cleaning",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/rooms/1205/command", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if len(writer.Frames) != 2 {
		t.Fatalf("expected 2 frames, got %d", len(writer.Frames))
	}

	gotSlave := writer.Frames[0]
	wantSlave := []byte{0x3a, 0x03, 0x20, 0x00, 0xa3}
	if !bytes.Equal(gotSlave, wantSlave) {
		t.Fatalf("unexpected slave frame: got % X want % X", gotSlave, wantSlave)
	}

	gotDisplay := writer.Frames[1]
	wantDisplay := []byte{0x3a, 0xfe, 0x35, 0x03}
	if !bytes.Equal(gotDisplay, wantDisplay) {
		t.Fatalf("unexpected display frame: got % X want % X", gotDisplay, wantDisplay)
	}
}

func TestCheckOutRouteBuildsLegacyFrames(t *testing.T) {
	server, writer := newTestServer()

	checkInBody, _ := json.Marshal(map[string]any{
		"request_id": "req-prep",
		"action":     "check_in",
		"stay_mode":  "overnight",
		"case":       "overnight",
	})
	checkInReq := httptest.NewRequest(http.MethodPost, "/api/rooms/1205/checkin", bytes.NewReader(checkInBody))
	checkInRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(checkInRec, checkInReq)
	if checkInRec.Code != http.StatusOK {
		t.Fatalf("expected prep checkin 200, got %d", checkInRec.Code)
	}

	writer.Frames = nil

	body, _ := json.Marshal(map[string]any{
		"request_id": "req-out",
		"action":     "check_out",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/rooms/1205/checkout", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if len(writer.Frames) != 2 {
		t.Fatalf("expected 2 frames, got %d", len(writer.Frames))
	}

	gotSlave := writer.Frames[0]
	wantSlave := []byte{0x3a, 0x03, 0x00, 0x00, 0xc3}
	if !bytes.Equal(gotSlave, wantSlave) {
		t.Fatalf("unexpected slave frame: got % X want % X", gotSlave, wantSlave)
	}

	gotDisplay := writer.Frames[1]
	wantDisplay := []byte{0x3a, 0xfe, 0x35, 0x00}
	if !bytes.Equal(gotDisplay, wantDisplay) {
		t.Fatalf("unexpected display frame: got % X want % X", gotDisplay, wantDisplay)
	}
}

func TestStateAndRoomDetailReflectMutations(t *testing.T) {
	server, _ := newTestServer()

	checkInBody, _ := json.Marshal(map[string]any{
		"request_id":     "req-3",
		"action":         "check_in",
		"stay_mode":      "temporary",
		"case":           "temporary",
		"stay_hours":     2,
		"payment_amount": 300.0,
		"payment_ref":    "POS-001",
	})
	checkInReq := httptest.NewRequest(http.MethodPost, "/api/rooms/1205/checkin", bytes.NewReader(checkInBody))
	checkInRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(checkInRec, checkInReq)
	if checkInRec.Code != http.StatusOK {
		t.Fatalf("expected checkin 200, got %d", checkInRec.Code)
	}

	stateReq := httptest.NewRequest(http.MethodGet, "/api/state", nil)
	stateRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(stateRec, stateReq)
	if stateRec.Code != http.StatusOK {
		t.Fatalf("expected state 200, got %d", stateRec.Code)
	}

	var stateResp struct {
		Rooms []map[string]any `json:"rooms"`
	}
	if err := json.NewDecoder(stateRec.Body).Decode(&stateResp); err != nil {
		t.Fatalf("decode state: %v", err)
	}
	if len(stateResp.Rooms) != 1 || stateResp.Rooms[0]["status"] != "occupied" {
		t.Fatalf("unexpected state rooms: %#v", stateResp.Rooms)
	}

	roomReq := httptest.NewRequest(http.MethodGet, "/api/rooms/1205", nil)
	roomRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(roomRec, roomReq)
	if roomRec.Code != http.StatusOK {
		t.Fatalf("expected room detail 200, got %d", roomRec.Code)
	}
}

func TestExtendAndPaymentRoutes(t *testing.T) {
	server, _ := newTestServer()

	extendBody, _ := json.Marshal(map[string]any{
		"request_id":     "req-4",
		"action":         "extend_stay",
		"stay_mode":      "temporary",
		"case":           "temporary",
		"extend_minutes": 60,
		"payment_amount": 150.0,
		"payment_ref":    "POS-002",
	})
	extendReq := httptest.NewRequest(http.MethodPost, "/api/rooms/1205/extend", bytes.NewReader(extendBody))
	extendRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(extendRec, extendReq)
	if extendRec.Code != http.StatusOK {
		t.Fatalf("expected extend 200, got %d", extendRec.Code)
	}

	paymentBody, _ := json.Marshal(map[string]any{
		"request_id": "req-5",
		"action":     "receive_payment",
		"amount":     150.0,
		"method":     "Cash",
		"reference":  "PAY-001",
	})
	paymentReq := httptest.NewRequest(http.MethodPost, "/api/rooms/1205/payment", bytes.NewReader(paymentBody))
	paymentRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(paymentRec, paymentReq)
	if paymentRec.Code != http.StatusOK {
		t.Fatalf("expected payment 200, got %d", paymentRec.Code)
	}
}

func TestAuditEventsRouteReturnsRecordedActions(t *testing.T) {
	server, _ := newTestServer()

	checkInBody, _ := json.Marshal(map[string]any{
		"request_id": "req-audit-1",
		"action":     "check_in",
		"stay_mode":  "overnight",
		"case":       "overnight",
	})
	checkInReq := httptest.NewRequest(http.MethodPost, "/api/rooms/1205/checkin", bytes.NewReader(checkInBody))
	checkInRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(checkInRec, checkInReq)
	if checkInRec.Code != http.StatusOK {
		t.Fatalf("expected checkin 200, got %d", checkInRec.Code)
	}

	auditReq := httptest.NewRequest(http.MethodGet, "/api/audit/events?room_id=1205", nil)
	auditRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(auditRec, auditReq)
	if auditRec.Code != http.StatusOK {
		t.Fatalf("expected audit 200, got %d", auditRec.Code)
	}

	var resp struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.NewDecoder(auditRec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode audit: %v", err)
	}
	if len(resp.Items) == 0 {
		t.Fatal("expected at least one audit item")
	}
	if resp.Items[0]["action"] != "check_in" {
		t.Fatalf("expected check_in action, got %#v", resp.Items[0]["action"])
	}
}

func TestOpenRoomAndServiceCommandsRecordAuditWithoutTransportFrames(t *testing.T) {
	server, writer := newTestServer()

	openBody, _ := json.Marshal(map[string]any{
		"request_id": "req-open",
		"action":     "open_room",
		"source":     "handheld",
		"actor":      "maid01",
		"reason":     "guest_request",
	})
	openReq := httptest.NewRequest(http.MethodPost, "/api/rooms/1205/command", bytes.NewReader(openBody))
	openRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(openRec, openReq)
	if openRec.Code != http.StatusOK {
		t.Fatalf("expected open_room 200, got %d", openRec.Code)
	}

	serviceBody, _ := json.Marshal(map[string]any{
		"request_id":   "req-service",
		"action":       "service",
		"source":       "handheld",
		"actor":        "maid01",
		"service_type": "towel",
	})
	serviceReq := httptest.NewRequest(http.MethodPost, "/api/rooms/1205/command", bytes.NewReader(serviceBody))
	serviceRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(serviceRec, serviceReq)
	if serviceRec.Code != http.StatusOK {
		t.Fatalf("expected service 200, got %d", serviceRec.Code)
	}

	if len(writer.Frames) != 0 {
		t.Fatalf("expected no transport frames for open_room/service, got %d", len(writer.Frames))
	}

	auditReq := httptest.NewRequest(http.MethodGet, "/api/audit/events?room_id=1205", nil)
	auditRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(auditRec, auditReq)
	if auditRec.Code != http.StatusOK {
		t.Fatalf("expected audit 200, got %d", auditRec.Code)
	}

	var resp struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.NewDecoder(auditRec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode audit: %v", err)
	}
	if len(resp.Items) < 2 {
		t.Fatalf("expected at least 2 audit items, got %d", len(resp.Items))
	}
	if resp.Items[0]["action"] != "service" || resp.Items[1]["action"] != "open_room" {
		t.Fatalf("unexpected audit order: %#v", resp.Items)
	}
}

func TestAlertsRouteReturnsVacantOpenRoomAlert(t *testing.T) {
	server, _ := newTestServer()

	openBody, _ := json.Marshal(map[string]any{
		"request_id": "req-alert-open",
		"action":     "open_room",
		"source":     "handheld",
		"actor":      "maid01",
	})
	openReq := httptest.NewRequest(http.MethodPost, "/api/rooms/1205/command", bytes.NewReader(openBody))
	openRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(openRec, openReq)
	if openRec.Code != http.StatusOK {
		t.Fatalf("expected open_room 200, got %d", openRec.Code)
	}

	alertReq := httptest.NewRequest(http.MethodGet, "/api/alerts", nil)
	alertRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(alertRec, alertReq)
	if alertRec.Code != http.StatusOK {
		t.Fatalf("expected alerts 200, got %d", alertRec.Code)
	}

	var resp struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.NewDecoder(alertRec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode alerts: %v", err)
	}
	if len(resp.Items) == 0 {
		t.Fatal("expected at least one alert item")
	}
	if resp.Items[0]["rule_code"] != "OPEN_ROOM_WHILE_VACANT" {
		t.Fatalf("expected OPEN_ROOM_WHILE_VACANT, got %#v", resp.Items[0]["rule_code"])
	}
}

func TestEventsStreamEmitsFullStateAndRoomUpdate(t *testing.T) {
	server, _ := newTestServer()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := httptest.NewRequest(http.MethodGet, "/events", nil).WithContext(ctx)
	req.Header.Set("Accept", "text/event-stream")

	stream := newStreamingRecorder()
	done := make(chan struct{})
	go func() {
		server.handleEvents(stream, req)
		close(done)
	}()

	waitForStreamContains(t, stream, "event: full_state")

	checkInBody, _ := json.Marshal(map[string]any{
		"request_id": "req-sse-1",
		"action":     "check_in",
		"stay_mode":  "overnight",
		"case":       "overnight",
	})
	checkInReq := httptest.NewRequest(http.MethodPost, "/api/rooms/1205/checkin", bytes.NewReader(checkInBody))
	checkInReq.Header.Set("Content-Type", "application/json")
	checkInRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(checkInRec, checkInReq)
	if checkInRec.Code != http.StatusOK {
		t.Fatalf("expected checkin 200, got %d", checkInRec.Code)
	}

	waitForStreamContains(t, stream, "event: room_update")
	waitForStreamContains(t, stream, `"request_id":"req-sse-1"`)
	waitForStreamContains(t, stream, `"status":"occupied"`)

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for SSE handler shutdown")
	}
}

func TestEventsStreamEmitsAuditAndFraudAlert(t *testing.T) {
	server, _ := newTestServer()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := httptest.NewRequest(http.MethodGet, "/events", nil).WithContext(ctx)
	req.Header.Set("Accept", "text/event-stream")

	stream := newStreamingRecorder()
	done := make(chan struct{})
	go func() {
		server.handleEvents(stream, req)
		close(done)
	}()

	waitForStreamContains(t, stream, "event: full_state")

	openBody, _ := json.Marshal(map[string]any{
		"request_id": "req-sse-open",
		"action":     "open_room",
		"source":     "handheld",
		"actor":      "maid01",
		"reason":     "guest_request",
	})
	openReq := httptest.NewRequest(http.MethodPost, "/api/rooms/1205/command", bytes.NewReader(openBody))
	openReq.Header.Set("Content-Type", "application/json")
	openRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(openRec, openReq)
	if openRec.Code != http.StatusOK {
		t.Fatalf("expected open_room 200, got %d", openRec.Code)
	}

	waitForStreamContains(t, stream, "event: audit_event")
	waitForStreamContains(t, stream, `"action":"open_room"`)
	waitForStreamContains(t, stream, "event: fraud_alert")
	waitForStreamContains(t, stream, `"rule_code":"OPEN_ROOM_WHILE_VACANT"`)

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for SSE handler shutdown")
	}
}

type streamingRecorder struct {
	mu     sync.Mutex
	header http.Header
	body   bytes.Buffer
	writes chan struct{}
	status int
}

func newStreamingRecorder() *streamingRecorder {
	return &streamingRecorder{
		header: make(http.Header),
		writes: make(chan struct{}, 16),
	}
}

func (r *streamingRecorder) Header() http.Header {
	return r.header
}

func (r *streamingRecorder) Write(data []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, err := r.body.Write(data)
	select {
	case r.writes <- struct{}{}:
	default:
	}
	return n, err
}

func (r *streamingRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
}

func (r *streamingRecorder) Flush() {}

func (r *streamingRecorder) String() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.body.String()
}

func waitForStreamContains(t *testing.T, stream *streamingRecorder, want string) {
	t.Helper()

	deadline := time.After(2 * time.Second)
	for {
		if strings.Contains(stream.String(), want) {
			return
		}
		select {
		case <-stream.writes:
		case <-deadline:
			t.Fatalf("timed out waiting for stream to contain %q; got %q", want, stream.String())
		}
	}
}
