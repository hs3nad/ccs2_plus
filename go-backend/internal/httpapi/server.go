package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"ccs2plus/go-backend/internal/legacy"
	"ccs2plus/go-backend/internal/roommap"
	backendstore "ccs2plus/go-backend/internal/store"
	"ccs2plus/go-backend/internal/transport"
)

type Server struct {
	transport *transport.Service
	roomRepo  roommap.Repository
	state     *stateStore
	events    *eventBroker
	store     *backendstore.Store
	dashboard string
}

func NewServer(transportSvc *transport.Service, roomRepo roommap.Repository) *Server {
	return NewServerWithStore(transportSvc, roomRepo, nil)
}

func NewServerWithStore(transportSvc *transport.Service, roomRepo roommap.Repository, store *backendstore.Store) *Server {
	state := newStateStore(roomRepo)
	if store != nil {
		if bootstrap, err := newStateStoreFromStore(roomRepo, store); err == nil {
			state = bootstrap
		}
	}
	return &Server{
		transport: transportSvc,
		roomRepo:  roomRepo,
		state:     state,
		events:    newEventBroker(),
		store:     store,
	}
}

func (s *Server) SetDashboardDir(path string) {
	s.dashboard = path
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", s.handlePing)
	mux.HandleFunc("/events", s.handleEvents)
	mux.HandleFunc("/api/state", s.handleState)
	mux.HandleFunc("/api/audit/events", s.handleAuditEvents)
	mux.HandleFunc("/api/alerts", s.handleAlerts)
	mux.HandleFunc("/api/config/gdrive", s.handleGoogleDriveConfig)
	mux.HandleFunc("/api/rooms/", s.handleRooms)
	if s.dashboard != "" {
		mux.HandleFunc("/", s.handleDashboard)
	}
	return withCORS(mux)
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.NotFound(w, r)
		return
	}

	path := filepath.Clean(strings.TrimPrefix(r.URL.Path, "/"))
	if path == "." {
		path = "index.html"
	}
	target := filepath.Join(s.dashboard, path)
	info, err := os.Stat(target)
	if err != nil || info.IsDir() {
		target = filepath.Join(s.dashboard, "index.html")
	}
	http.ServeFile(w, r, target)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
		w.Header().Set("Access-Control-Max-Age", "86400")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type checkInRequest struct {
	RequestID     string  `json:"request_id"`
	Action        string  `json:"action"`
	StayMode      string  `json:"stay_mode"`
	Case          string  `json:"case"`
	StayHours     int     `json:"stay_hours"`
	PaymentAmount float64 `json:"payment_amount"`
	PaymentRef    string  `json:"payment_ref"`
}

type checkoutRequest struct {
	RequestID string `json:"request_id"`
	Action    string `json:"action"`
}

type roomCommandRequest struct {
	RequestID   string `json:"request_id"`
	Action      string `json:"action"`
	Source      string `json:"source"`
	Actor       string `json:"actor"`
	Reason      string `json:"reason"`
	ServiceType string `json:"service_type"`
}

type extendRequest struct {
	RequestID     string  `json:"request_id"`
	Action        string  `json:"action"`
	StayMode      string  `json:"stay_mode"`
	Case          string  `json:"case"`
	ExtendMinutes int     `json:"extend_minutes"`
	PaymentAmount float64 `json:"payment_amount"`
	PaymentRef    string  `json:"payment_ref"`
}

type paymentRequest struct {
	RequestID string  `json:"request_id"`
	Action    string  `json:"action"`
	Amount    float64 `json:"amount"`
	Method    string  `json:"method"`
	Reference string  `json:"reference"`
}

type googleDriveConfigRequest struct {
	CredentialsPath string `json:"credentials_path"`
	FolderID        string `json:"folder_id"`
	Enabled         bool   `json:"enabled"`
}

type roomState struct {
	RoomID            string    `json:"room_id"`
	Floor             int       `json:"floor"`
	Status            string    `json:"status"`
	StayMode          string    `json:"stay_mode"`
	CaseCode          string    `json:"case"`
	ControllerOnline  bool      `json:"controller_online"`
	RemainingMinutes  int       `json:"remaining_minutes"`
	LastPaymentAmount float64   `json:"last_payment_amount,omitempty"`
	LastPaymentMethod string    `json:"last_payment_method,omitempty"`
	LastPaymentRef    string    `json:"last_payment_ref,omitempty"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type roomSnapshot struct {
	RoomID           string `json:"room_id"`
	Floor            int    `json:"floor"`
	Status           string `json:"status"`
	StayMode         string `json:"stay_mode"`
	CaseCode         string `json:"case"`
	ControllerOnline bool   `json:"controller_online"`
	RemainingMinutes *int   `json:"remaining_minutes,omitempty"`
}

type stateStore struct {
	mu        sync.RWMutex
	rooms     map[string]roomState
	audit     []auditEvent
	nextAudit int
}

type auditEvent struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	RoomID    string    `json:"room_id"`
	Action    string    `json:"action"`
	Actor     string    `json:"actor"`
	Result    string    `json:"result"`
	StayMode  string    `json:"stay_mode,omitempty"`
	CaseCode  string    `json:"case,omitempty"`
	Detail    string    `json:"detail,omitempty"`
}

type alertItem struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	RoomID    string    `json:"room_id"`
	Severity  string    `json:"severity"`
	RuleCode  string    `json:"rule_code"`
	Status    string    `json:"status"`
	Title     string    `json:"title"`
	Detail    string    `json:"detail"`
}

func newStateStore(repo roommap.Repository) *stateStore {
	rooms := make(map[string]roomState)
	for _, mapping := range repo.List() {
		rooms[mapping.RoomID] = roomState{
			RoomID:           mapping.RoomID,
			Floor:            floorFromRoomID(mapping.RoomID),
			Status:           "vacant",
			StayMode:         "",
			CaseCode:         "",
			ControllerOnline: true,
			UpdatedAt:        time.Now(),
		}
	}
	return &stateStore{rooms: rooms}
}

func newStateStoreFromStore(repo roommap.Repository, store *backendstore.Store) (*stateStore, error) {
	items, audits, err := store.Bootstrap()
	if err != nil {
		return nil, err
	}

	state := newStateStore(repo)
	state.mu.Lock()
	defer state.mu.Unlock()

	for _, item := range items {
		state.rooms[item.RoomID] = roomState{
			RoomID:            item.RoomID,
			Floor:             item.Floor,
			Status:            item.Status,
			StayMode:          item.StayMode,
			CaseCode:          item.CaseCode,
			ControllerOnline:  item.ControllerOnline,
			RemainingMinutes:  item.RemainingMinutes,
			LastPaymentAmount: item.LastPaymentAmount,
			LastPaymentMethod: item.LastPaymentMethod,
			LastPaymentRef:    item.LastPaymentRef,
			UpdatedAt:         item.UpdatedAt,
		}
	}

	state.audit = make([]auditEvent, 0, len(audits))
	for _, item := range audits {
		state.audit = append(state.audit, auditEvent{
			ID:        item.ID,
			Timestamp: item.Timestamp,
			RoomID:    item.RoomID,
			Action:    item.Action,
			Actor:     item.Actor,
			Result:    item.Result,
			StayMode:  item.StayMode,
			CaseCode:  item.CaseCode,
			Detail:    item.Detail,
		})
	}
	state.nextAudit = len(audits)
	return state, nil
}

func (s *Server) handlePing(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	items := s.state.list()
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

	writeJSON(w, http.StatusOK, map[string]any{
		"rooms":   rooms,
		"summary": summary,
	})
}

func (s *Server) handleAuditEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items": s.state.listAudit(r.URL.Query().Get("room_id")),
	})
}

func (s *Server) handleAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items": s.state.buildAlerts(),
	})
}

func (s *Server) handleGoogleDriveConfig(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "CONFIG_STORE_UNAVAILABLE", "config store is not available", "")
		return
	}

	switch r.Method {
	case http.MethodGet:
		cfg, err := s.store.GetGoogleDriveConfig()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "CONFIG_READ_FAILED", err.Error(), "")
			return
		}
		writeJSON(w, http.StatusOK, googleDriveConfigPayload(cfg))
	case http.MethodPut:
		var req googleDriveConfigRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error(), "")
			return
		}
		if req.Enabled && (strings.TrimSpace(req.CredentialsPath) == "" || strings.TrimSpace(req.FolderID) == "") {
			writeError(w, http.StatusBadRequest, "INVALID_GDRIVE_CONFIG", "credentials_path and folder_id are required when enabled", "")
			return
		}
		cfg, err := s.store.SaveGoogleDriveConfig(backendstore.GoogleDriveConfig{
			CredentialsPath: strings.TrimSpace(req.CredentialsPath),
			FolderID:        strings.TrimSpace(req.FolderID),
			Enabled:         req.Enabled,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "CONFIG_WRITE_FAILED", err.Error(), "")
			return
		}
		writeJSON(w, http.StatusOK, googleDriveConfigPayload(cfg))
	default:
		http.NotFound(w, r)
	}
}

func googleDriveConfigPayload(cfg backendstore.GoogleDriveConfig) map[string]any {
	updatedAt := ""
	if !cfg.UpdatedAt.IsZero() {
		updatedAt = cfg.UpdatedAt.Format(time.RFC3339)
	}
	return map[string]any{
		"credentials_path": cfg.CredentialsPath,
		"folder_id":        cfg.FolderID,
		"enabled":          cfg.Enabled,
		"updated_at":       updatedAt,
		"configured":       cfg.CredentialsPath != "" && cfg.FolderID != "",
	}
}

func (s *Server) handleRooms(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/rooms/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 1 && r.Method == http.MethodGet {
		s.handleRoomDetail(w, r, parts[0])
		return
	}
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}

	roomID := parts[0]
	action := parts[1]

	switch {
	case r.Method == http.MethodPost && action == "checkin":
		s.handleCheckIn(w, r, roomID)
	case r.Method == http.MethodPost && action == "checkout":
		s.handleCheckOut(w, r, roomID)
	case r.Method == http.MethodPost && action == "extend":
		s.handleExtend(w, r, roomID)
	case r.Method == http.MethodPost && action == "payment":
		s.handlePayment(w, r, roomID)
	case r.Method == http.MethodPost && action == "command":
		s.handleRoomCommand(w, r, roomID)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleCheckIn(w http.ResponseWriter, r *http.Request, roomID string) {
	var req checkInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error(), "")
		return
	}

	result, err := s.apply(r.Context(), transport.CommandIntent{
		RequestID:   requestIDOrNow(req.RequestID),
		RoomID:      roomID,
		Action:      "check_in",
		TargetState: legacy.SemanticOccupied,
		StayMode:    req.StayMode,
		CaseCode:    fallbackCase(req.Case, req.StayMode),
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, "TRANSPORT_ERROR", err.Error(), req.RequestID)
		return
	}

	requestID := result["request_id"].(string)
	caseCode := fallbackCase(req.Case, req.StayMode)
	now := time.Now()
	remainingMinutes := maxInt(req.StayHours*60, defaultMinutesForStayMode(req.StayMode))
	if s.store != nil {
		room, ok := s.state.get(roomID)
		if !ok {
			writeError(w, http.StatusNotFound, "ROOM_NOT_FOUND", "room not found", requestID)
			return
		}
		if err := s.store.RecordCheckIn(r.Context(), backendstore.CheckInParams{
			RequestID:        requestID,
			RoomID:           roomID,
			Floor:            room.Floor,
			StayMode:         req.StayMode,
			CaseCode:         caseCode,
			RemainingMinutes: remainingMinutes,
			PaymentAmount:    req.PaymentAmount,
			PaymentRef:       req.PaymentRef,
			Actor:            "cashier",
			Timestamp:        now,
		}); err != nil {
			writeError(w, http.StatusInternalServerError, "DB_WRITE_FAILED", err.Error(), requestID)
			return
		}
	}

	s.state.applyCheckIn(roomID, req.StayMode, caseCode, req.StayHours, req.PaymentAmount, req.PaymentRef)
	audit := s.state.appendAudit(auditEvent{
		ID:        requestID,
		Timestamp: now,
		RoomID:    roomID,
		Action:    "check_in",
		Actor:     "cashier",
		Result:    "accepted",
		StayMode:  req.StayMode,
		CaseCode:  caseCode,
		Detail:    req.PaymentRef,
	})
	s.publishAuditEvent(audit)
	s.publishRoomUpdate(roomID, requestID, "check_in")
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleCheckOut(w http.ResponseWriter, r *http.Request, roomID string) {
	var req checkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error(), "")
		return
	}

	result, err := s.apply(r.Context(), transport.CommandIntent{
		RequestID:   requestIDOrNow(req.RequestID),
		RoomID:      roomID,
		Action:      "check_out",
		TargetState: legacy.SemanticVacant,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, "TRANSPORT_ERROR", err.Error(), req.RequestID)
		return
	}

	requestID := result["request_id"].(string)
	now := time.Now()
	if s.store != nil {
		if err := s.store.RecordCheckOut(r.Context(), backendstore.CheckOutParams{
			RequestID: requestID,
			RoomID:    roomID,
			Actor:     "cashier",
			Timestamp: now,
		}); err != nil {
			writeError(w, http.StatusInternalServerError, "DB_WRITE_FAILED", err.Error(), requestID)
			return
		}
	}

	s.state.applyCheckOut(roomID)
	audit := s.state.appendAudit(auditEvent{
		ID:        requestID,
		Timestamp: now,
		RoomID:    roomID,
		Action:    "check_out",
		Actor:     "cashier",
		Result:    "accepted",
	})
	s.publishAuditEvent(audit)
	s.publishRoomUpdate(roomID, requestID, "check_out")
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleRoomCommand(w http.ResponseWriter, r *http.Request, roomID string) {
	var req roomCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error(), "")
		return
	}

	intent, err := commandIntentFromAction(roomID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ACTION", err.Error(), req.RequestID)
		return
	}

	if req.Action == "open_room" || req.Action == "service" {
		requestID := requestIDOrNow(req.RequestID)
		now := time.Now()
		detail := fallbackDetail(req.Reason, req.ServiceType)
		actor := fallbackActor(req.Actor, req.Source)
		if s.store != nil {
			if err := s.store.RecordCommand(r.Context(), backendstore.CommandParams{
				RequestID: requestID,
				RoomID:    roomID,
				Action:    req.Action,
				Actor:     actor,
				Detail:    detail,
				Timestamp: now,
			}); err != nil {
				writeError(w, http.StatusInternalServerError, "DB_WRITE_FAILED", err.Error(), requestID)
				return
			}
		}
		s.state.applyCommand(roomID, req.Action)
		audit := s.state.appendAudit(auditEvent{
			ID:        requestID,
			Timestamp: now,
			RoomID:    roomID,
			Action:    req.Action,
			Actor:     actor,
			Result:    "accepted",
			Detail:    detail,
		})
		s.publishAuditEvent(audit)
		if alert, ok := s.state.alertForAudit(audit); ok {
			s.publishFraudAlert(alert)
		}
		s.publishRoomUpdate(roomID, requestID, req.Action)
		writeJSON(w, http.StatusOK, map[string]any{
			"request_id": requestID,
			"room_id":    roomID,
			"action":     req.Action,
			"status":     "accepted",
			"detail":     detail,
			"sent_at":    now.Format(time.RFC3339),
		})
		return
	}

	result, err := s.apply(r.Context(), intent)
	if err != nil {
		writeError(w, http.StatusBadRequest, "TRANSPORT_ERROR", err.Error(), req.RequestID)
		return
	}

	requestID := result["request_id"].(string)
	now := time.Now()
	actor := fallbackActor(req.Actor, req.Source)
	if s.store != nil {
		if err := s.store.RecordCommand(r.Context(), backendstore.CommandParams{
			RequestID: requestID,
			RoomID:    roomID,
			Action:    req.Action,
			Actor:     actor,
			Timestamp: now,
		}); err != nil {
			writeError(w, http.StatusInternalServerError, "DB_WRITE_FAILED", err.Error(), requestID)
			return
		}
	}
	s.state.applyCommand(roomID, req.Action)
	audit := s.state.appendAudit(auditEvent{
		ID:        requestID,
		Timestamp: now,
		RoomID:    roomID,
		Action:    req.Action,
		Actor:     actor,
		Result:    "accepted",
	})
	s.publishAuditEvent(audit)
	s.publishRoomUpdate(roomID, requestID, req.Action)
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleExtend(w http.ResponseWriter, r *http.Request, roomID string) {
	var req extendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error(), "")
		return
	}

	intent := transport.CommandIntent{
		RequestID:   requestIDOrNow(req.RequestID),
		RoomID:      roomID,
		Action:      "extend_stay",
		TargetState: legacy.SemanticOccupied,
		StayMode:    req.StayMode,
		CaseCode:    fallbackCase(req.Case, req.StayMode),
	}
	result, err := s.apply(r.Context(), intent)
	if err != nil {
		writeError(w, http.StatusBadRequest, "TRANSPORT_ERROR", err.Error(), req.RequestID)
		return
	}

	requestID := result["request_id"].(string)
	caseCode := fallbackCase(req.Case, req.StayMode)
	now := time.Now()
	current, _ := s.state.get(roomID)
	remainingMinutes := current.RemainingMinutes + req.ExtendMinutes
	if remainingMinutes < req.ExtendMinutes {
		remainingMinutes = req.ExtendMinutes
	}
	if s.store != nil {
		if err := s.store.RecordExtend(r.Context(), backendstore.ExtendParams{
			RequestID:        requestID,
			RoomID:           roomID,
			StayMode:         req.StayMode,
			CaseCode:         caseCode,
			RemainingMinutes: remainingMinutes,
			ExtendMinutes:    req.ExtendMinutes,
			PaymentAmount:    req.PaymentAmount,
			PaymentRef:       req.PaymentRef,
			Actor:            "cashier",
			Timestamp:        now,
		}); err != nil {
			writeError(w, http.StatusInternalServerError, "DB_WRITE_FAILED", err.Error(), requestID)
			return
		}
	}

	s.state.applyExtend(roomID, req.StayMode, caseCode, req.ExtendMinutes, req.PaymentAmount, req.PaymentRef)
	audit := s.state.appendAudit(auditEvent{
		ID:        requestID,
		Timestamp: now,
		RoomID:    roomID,
		Action:    "extend_stay",
		Actor:     "cashier",
		Result:    "accepted",
		StayMode:  req.StayMode,
		CaseCode:  caseCode,
		Detail:    req.PaymentRef,
	})
	s.publishAuditEvent(audit)
	s.publishRoomUpdate(roomID, requestID, "extend_stay")
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handlePayment(w http.ResponseWriter, r *http.Request, roomID string) {
	var req paymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error(), "")
		return
	}

	room, ok := s.state.get(roomID)
	if !ok {
		writeError(w, http.StatusNotFound, "ROOM_NOT_FOUND", "room not found", req.RequestID)
		return
	}

	requestID := requestIDOrNow(req.RequestID)
	now := time.Now()
	if s.store != nil {
		if err := s.store.RecordPayment(r.Context(), backendstore.PaymentParams{
			RequestID: requestID,
			RoomID:    roomID,
			Amount:    req.Amount,
			Method:    req.Method,
			Reference: req.Reference,
			Actor:     "cashier",
			Timestamp: now,
		}); err != nil {
			writeError(w, http.StatusInternalServerError, "DB_WRITE_FAILED", err.Error(), requestID)
			return
		}
	}
	s.state.applyPayment(roomID, req.Amount, req.Method, req.Reference)
	audit := s.state.appendAudit(auditEvent{
		ID:        requestID,
		Timestamp: now,
		RoomID:    roomID,
		Action:    "receive_payment",
		Actor:     "cashier",
		Result:    "accepted",
		StayMode:  room.StayMode,
		CaseCode:  room.CaseCode,
		Detail:    req.Reference,
	})
	s.publishAuditEvent(audit)
	s.publishRoomUpdate(roomID, requestID, "receive_payment")
	writeJSON(w, http.StatusOK, map[string]any{
		"request_id": requestID,
		"room_id":    roomID,
		"action":     "receive_payment",
		"status":     "accepted",
		"amount":     req.Amount,
		"method":     req.Method,
		"reference":  req.Reference,
		"stay_mode":  room.StayMode,
		"case":       room.CaseCode,
		"paid_at":    now.Format(time.RFC3339),
	})
}

func (s *Server) handleRoomDetail(w http.ResponseWriter, r *http.Request, roomID string) {
	room, ok := s.state.get(roomID)
	if !ok {
		http.NotFound(w, r)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"room_id":           room.RoomID,
		"floor":             room.Floor,
		"status":            room.Status,
		"stay_mode":         room.StayMode,
		"case":              room.CaseCode,
		"controller_online": room.ControllerOnline,
		"remaining_minutes": room.RemainingMinutes,
		"last_payment": map[string]any{
			"amount":    room.LastPaymentAmount,
			"method":    room.LastPaymentMethod,
			"reference": room.LastPaymentRef,
		},
		"updated_at": room.UpdatedAt.Format(time.RFC3339),
	})
}

func commandIntentFromAction(roomID string, req roomCommandRequest) (transport.CommandIntent, error) {
	intent := transport.CommandIntent{
		RequestID: requestIDOrNow(req.RequestID),
		RoomID:    roomID,
		Action:    req.Action,
	}

	switch req.Action {
	case "start_cleaning":
		intent.TargetState = legacy.SemanticCleaning
	case "finish_cleaning":
		intent.TargetState = legacy.SemanticVacant
	case "open_room", "service":
		intent.TargetState = legacy.SemanticOccupied
	default:
		return transport.CommandIntent{}, fmt.Errorf("unsupported command action: %s", req.Action)
	}

	return intent, nil
}

func (s *Server) apply(ctx context.Context, intent transport.CommandIntent) (map[string]any, error) {
	result, err := s.transport.ApplyIntent(ctx, intent)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"request_id":    result.RequestID,
		"room_id":       result.RoomID,
		"action":        result.Action,
		"status":        "accepted",
		"stay_mode":     result.StayMode,
		"case":          result.CaseCode,
		"target_state":  result.TargetState,
		"slave_frame":   fmt.Sprintf("% X", result.SlaveFrame),
		"display_frame": fmt.Sprintf("% X", result.DisplayFrame),
		"sent_at":       result.SentAt.Format(time.RFC3339),
	}, nil
}

func requestIDOrNow(requestID string) string {
	if requestID != "" {
		return requestID
	}
	return fmt.Sprintf("req-%d", time.Now().UnixNano())
}

func fallbackCase(caseCode, stayMode string) string {
	if caseCode != "" {
		return caseCode
	}
	return stayMode
}

func fallbackActor(actor, source string) string {
	if actor != "" {
		return actor
	}
	if source != "" {
		return source
	}
	return "system"
}

func fallbackDetail(reason, serviceType string) string {
	if serviceType != "" {
		return serviceType
	}
	return reason
}

func (s *stateStore) list() []roomState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]roomState, 0, len(s.rooms))
	for _, item := range s.rooms {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].RoomID < items[j].RoomID
	})
	return items
}

func (s *stateStore) appendAudit(entry auditEvent) auditEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextAudit++
	if entry.ID == "" {
		entry.ID = fmt.Sprintf("audit-%d", s.nextAudit)
	}
	s.audit = append([]auditEvent{entry}, s.audit...)
	if len(s.audit) > 200 {
		s.audit = s.audit[:200]
	}
	return entry
}

func (s *stateStore) listAudit(roomID string) []auditEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]auditEvent, 0, len(s.audit))
	for _, item := range s.audit {
		if roomID != "" && item.RoomID != roomID {
			continue
		}
		items = append(items, item)
	}
	return items
}

func (s *stateStore) buildAlerts() []alertItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]alertItem, 0)
	for _, room := range s.rooms {
		if !room.ControllerOnline {
			items = append(items, alertItem{
				ID:        "alert-offline-" + room.RoomID,
				Timestamp: room.UpdatedAt,
				RoomID:    room.RoomID,
				Severity:  "high",
				RuleCode:  "DEVICE_OFFLINE",
				Status:    "open",
				Title:     "Offline device",
				Detail:    "Controller is offline while room remains in the active system map.",
			})
		}
		if room.Status == "overstay" {
			items = append(items, alertItem{
				ID:        "alert-overstay-" + room.RoomID,
				Timestamp: room.UpdatedAt,
				RoomID:    room.RoomID,
				Severity:  "medium",
				RuleCode:  "POSSIBLE_EXTEND_WITHOUT_PAYMENT",
				Status:    "open",
				Title:     "Possible extend without payment",
				Detail:    "Room remains overstay and should be reviewed by cashier or manager.",
			})
		}
	}
	for _, audit := range s.audit {
		if audit.Action == "open_room" {
			room := s.rooms[audit.RoomID]
			if room.Status == "vacant" {
				items = append(items, alertItem{
					ID:        "alert-open-room-" + audit.ID,
					Timestamp: audit.Timestamp,
					RoomID:    audit.RoomID,
					Severity:  "high",
					RuleCode:  "OPEN_ROOM_WHILE_VACANT",
					Status:    "open",
					Title:     "Open room while vacant",
					Detail:    "Open-room event was recorded while room state remained vacant.",
				})
			}
		}
	}
	return items
}

func (s *stateStore) get(roomID string) (roomState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.rooms[roomID]
	return item, ok
}

func (s *stateStore) applyCheckIn(roomID, stayMode, caseCode string, stayHours int, paymentAmount float64, paymentRef string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.rooms[roomID]
	if !ok {
		return
	}
	item.Status = "occupied"
	item.StayMode = stayMode
	item.CaseCode = caseCode
	item.RemainingMinutes = maxInt(stayHours*60, defaultMinutesForStayMode(stayMode))
	if paymentAmount > 0 {
		item.LastPaymentAmount = paymentAmount
		item.LastPaymentRef = paymentRef
		item.LastPaymentMethod = "check_in"
	}
	item.UpdatedAt = time.Now()
	s.rooms[roomID] = item
}

func (s *stateStore) applyCheckOut(roomID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.rooms[roomID]
	if !ok {
		return
	}
	item.Status = "vacant"
	item.StayMode = ""
	item.CaseCode = ""
	item.RemainingMinutes = 0
	item.UpdatedAt = time.Now()
	s.rooms[roomID] = item
}

func (s *stateStore) applyCommand(roomID, action string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.rooms[roomID]
	if !ok {
		return
	}
	switch action {
	case "start_cleaning":
		item.Status = "cleaning"
	case "finish_cleaning":
		item.Status = "vacant"
		item.StayMode = ""
		item.CaseCode = ""
		item.RemainingMinutes = 0
	}
	item.UpdatedAt = time.Now()
	s.rooms[roomID] = item
}

func (s *stateStore) applyExtend(roomID, stayMode, caseCode string, extendMinutes int, paymentAmount float64, paymentRef string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.rooms[roomID]
	if !ok {
		return
	}
	item.Status = "occupied"
	if stayMode != "" {
		item.StayMode = stayMode
	}
	if caseCode != "" {
		item.CaseCode = caseCode
	}
	item.RemainingMinutes += extendMinutes
	if paymentAmount > 0 {
		item.LastPaymentAmount = paymentAmount
		item.LastPaymentRef = paymentRef
		item.LastPaymentMethod = "extend"
	}
	item.UpdatedAt = time.Now()
	s.rooms[roomID] = item
}

func (s *stateStore) applyPayment(roomID string, amount float64, method, ref string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.rooms[roomID]
	if !ok {
		return
	}
	item.LastPaymentAmount = amount
	item.LastPaymentMethod = method
	item.LastPaymentRef = ref
	item.UpdatedAt = time.Now()
	s.rooms[roomID] = item
}

func (r roomState) snapshot() roomSnapshot {
	snapshot := roomSnapshot{
		RoomID:           r.RoomID,
		Floor:            r.Floor,
		Status:           r.Status,
		StayMode:         r.StayMode,
		CaseCode:         r.CaseCode,
		ControllerOnline: r.ControllerOnline,
	}
	if r.RemainingMinutes > 0 {
		value := r.RemainingMinutes
		snapshot.RemainingMinutes = &value
	}
	return snapshot
}

func floorFromRoomID(roomID string) int {
	if len(roomID) == 0 {
		return 0
	}
	if len(roomID) == 1 {
		return int(roomID[0] - '0')
	}
	return int(roomID[0] - '0')
}

func defaultMinutesForStayMode(stayMode string) int {
	switch stayMode {
	case "temporary":
		return 120
	case "monthly":
		return 43200
	case "yearly":
		return 525600
	default:
		return 720
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code, message, requestID string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]any{
			"code":       code,
			"message":    message,
			"request_id": requestID,
		},
	})
}

func WithTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, 3*time.Second)
}
