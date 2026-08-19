package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"ccs2plus/go-backend/internal/roommap"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

type RoomRecord struct {
	RoomID            string
	Floor             int
	Status            string
	StayMode          string
	CaseCode          string
	ControllerOnline  bool
	RemainingMinutes  int
	LastPaymentAmount float64
	LastPaymentMethod string
	LastPaymentRef    string
	UpdatedAt         time.Time
}

type AuditRecord struct {
	ID        string
	Timestamp time.Time
	RoomID    string
	Action    string
	Actor     string
	Result    string
	StayMode  string
	CaseCode  string
	Detail    string
}

type StayRecord struct {
	StayID           string
	RequestID        string
	RoomID           string
	StayMode         string
	CaseCode         string
	Status           string
	GuestInAt        *time.Time
	GuestOutAt       *time.Time
	CleanUpAt        *time.Time
	AllocatedMinutes int
	UsedMinutes      int
	PaymentAmount    float64
	PaymentRef       string
	Remark           string
	LastUpdatedAt    time.Time
}

type PaymentRecord struct {
	PaymentID string
	RequestID string
	StayID    string
	RoomID    string
	Amount    float64
	Method    string
	Reference string
	PaidAt    time.Time
}

type ExportRunRecord struct {
	ExportDate  string
	Status      string
	LocalPath   string
	DriveFileID string
	LastError   string
	UploadedAt  *time.Time
	UpdatedAt   time.Time
}

type GoogleDriveConfig struct {
	CredentialsPath string
	FolderID        string
	Enabled         bool
	UpdatedAt       time.Time
}

type CheckInParams struct {
	RequestID        string
	RoomID           string
	Floor            int
	StayMode         string
	CaseCode         string
	RemainingMinutes int
	PaymentAmount    float64
	PaymentRef       string
	Actor            string
	Timestamp        time.Time
}

type CheckOutParams struct {
	RequestID string
	RoomID    string
	Actor     string
	Timestamp time.Time
}

type ExtendParams struct {
	RequestID        string
	RoomID           string
	StayMode         string
	CaseCode         string
	RemainingMinutes int
	ExtendMinutes    int
	PaymentAmount    float64
	PaymentRef       string
	Actor            string
	Timestamp        time.Time
}

type PaymentParams struct {
	RequestID string
	RoomID    string
	Amount    float64
	Method    string
	Reference string
	Actor     string
	Timestamp time.Time
}

type CommandParams struct {
	RequestID string
	RoomID    string
	Action    string
	Actor     string
	Detail    string
	Timestamp time.Time
}

type DailyWorkbookData struct {
	Date     string
	Rooms    []RoomRecord
	Stays    []StayRecord
	Payments []PaymentRecord
	Audits   []AuditRecord
}

type MonthlyWorkbookData struct {
	MonthKey string
	Rooms    []RoomRecord
	Stays    []StayRecord
	Payments []PaymentRecord
	Audits   []AuditRecord
}

func Open(path string, repo roommap.Repository) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)

	store := &Store{db: db}
	if err := store.exec(`PRAGMA foreign_keys = ON;`); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.exec(`PRAGMA busy_timeout = 5000;`); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.seedRooms(repo.List()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) exec(query string) error {
	if _, err := s.db.Exec(query); err != nil {
		return fmt.Errorf("exec query: %w", err)
	}
	return nil
}

func (s *Store) migrate() error {
	schema := []string{
		`CREATE TABLE IF NOT EXISTS rooms (
			room_id TEXT PRIMARY KEY,
			floor INTEGER NOT NULL,
			controller_id TEXT NOT NULL,
			rs485_address INTEGER NOT NULL,
			card_index INTEGER NOT NULL,
			port_index INTEGER NOT NULL,
			display_enabled INTEGER NOT NULL DEFAULT 1,
			remark TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'vacant',
			stay_mode TEXT NOT NULL DEFAULT '',
			case_code TEXT NOT NULL DEFAULT '',
			controller_online INTEGER NOT NULL DEFAULT 1,
			remaining_minutes INTEGER NOT NULL DEFAULT 0,
			last_payment_amount REAL NOT NULL DEFAULT 0,
			last_payment_method TEXT NOT NULL DEFAULT '',
			last_payment_ref TEXT NOT NULL DEFAULT '',
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS stays (
			stay_id TEXT PRIMARY KEY,
			request_id TEXT NOT NULL,
			room_id TEXT NOT NULL,
			stay_mode TEXT NOT NULL,
			case_code TEXT NOT NULL,
			status TEXT NOT NULL,
			guest_in_at TEXT,
			guest_out_at TEXT,
			clean_up_at TEXT,
			allocated_minutes INTEGER NOT NULL DEFAULT 0,
			used_minutes INTEGER NOT NULL DEFAULT 0,
			payment_amount REAL NOT NULL DEFAULT 0,
			payment_ref TEXT NOT NULL DEFAULT '',
			remark TEXT NOT NULL DEFAULT '',
			last_updated_at TEXT NOT NULL,
			FOREIGN KEY(room_id) REFERENCES rooms(room_id)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_stays_room_status ON stays(room_id, status);`,
		`CREATE TABLE IF NOT EXISTS payments (
			payment_id TEXT PRIMARY KEY,
			request_id TEXT NOT NULL,
			stay_id TEXT NOT NULL DEFAULT '',
			room_id TEXT NOT NULL,
			amount REAL NOT NULL,
			method TEXT NOT NULL,
			reference TEXT NOT NULL,
			paid_at TEXT NOT NULL,
			FOREIGN KEY(room_id) REFERENCES rooms(room_id)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_payments_room_paid_at ON payments(room_id, paid_at);`,
		`CREATE TABLE IF NOT EXISTS audit_events (
			event_id TEXT PRIMARY KEY,
			room_id TEXT NOT NULL,
			action TEXT NOT NULL,
			actor TEXT NOT NULL,
			result TEXT NOT NULL,
			stay_mode TEXT NOT NULL DEFAULT '',
			case_code TEXT NOT NULL DEFAULT '',
			detail TEXT NOT NULL DEFAULT '',
			event_time TEXT NOT NULL,
			FOREIGN KEY(room_id) REFERENCES rooms(room_id)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_audit_events_room_time ON audit_events(room_id, event_time DESC);`,
		`CREATE TABLE IF NOT EXISTS export_runs (
			export_date TEXT PRIMARY KEY,
			status TEXT NOT NULL,
			local_path TEXT NOT NULL DEFAULT '',
			drive_file_id TEXT NOT NULL DEFAULT '',
			last_error TEXT NOT NULL DEFAULT '',
			uploaded_at TEXT,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS report_runs (
			report_kind TEXT NOT NULL,
			period_key TEXT NOT NULL,
			status TEXT NOT NULL,
			local_path TEXT NOT NULL DEFAULT '',
			drive_file_id TEXT NOT NULL DEFAULT '',
			last_error TEXT NOT NULL DEFAULT '',
			uploaded_at TEXT,
			updated_at TEXT NOT NULL,
			PRIMARY KEY(report_kind, period_key)
		);`,
		`CREATE TABLE IF NOT EXISTS app_settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
	}
	for _, query := range schema {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("migrate sqlite schema: %w", err)
		}
	}
	return nil
}

func (s *Store) seedRooms(mappings []roommap.RoomMapping) error {
	now := time.Now().Format(time.RFC3339)
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin room seed tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	stmt, err := tx.Prepare(`
		INSERT INTO rooms (
			room_id, floor, controller_id, rs485_address, card_index, port_index, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(room_id) DO UPDATE SET
			floor = excluded.floor,
			controller_id = excluded.controller_id,
			rs485_address = excluded.rs485_address,
			card_index = excluded.card_index,
			port_index = excluded.port_index
	`)
	if err != nil {
		return fmt.Errorf("prepare room seed: %w", err)
	}
	defer stmt.Close()

	for _, mapping := range mappings {
		if _, err = stmt.Exec(
			mapping.RoomID,
			floorFromRoomID(mapping.RoomID),
			mapping.ControllerID,
			mapping.Address,
			mapping.CardIndex,
			mapping.PortIndex,
			now,
		); err != nil {
			return fmt.Errorf("seed room %s: %w", mapping.RoomID, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit room seed: %w", err)
	}
	return nil
}

func (s *Store) Bootstrap() ([]RoomRecord, []AuditRecord, error) {
	rooms, err := s.ListRooms()
	if err != nil {
		return nil, nil, err
	}
	audits, err := s.ListRecentAudit("", 200)
	if err != nil {
		return nil, nil, err
	}
	return rooms, audits, nil
}

func (s *Store) GetGoogleDriveConfig() (GoogleDriveConfig, error) {
	var raw string
	var updatedAt string
	err := s.db.QueryRow(`
		SELECT value, updated_at
		FROM app_settings
		WHERE key = 'google_drive'
	`).Scan(&raw, &updatedAt)
	if err == sql.ErrNoRows {
		return GoogleDriveConfig{}, nil
	}
	if err != nil {
		return GoogleDriveConfig{}, fmt.Errorf("load google drive config: %w", err)
	}

	var cfg GoogleDriveConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return GoogleDriveConfig{}, fmt.Errorf("decode google drive config: %w", err)
	}
	if updatedAt != "" {
		if parsed, err := time.Parse(time.RFC3339, updatedAt); err == nil {
			cfg.UpdatedAt = parsed
		}
	}
	return cfg, nil
}

func (s *Store) SaveGoogleDriveConfig(cfg GoogleDriveConfig) (GoogleDriveConfig, error) {
	now := time.Now()
	cfg.UpdatedAt = now
	payload, err := json.Marshal(cfg)
	if err != nil {
		return GoogleDriveConfig{}, fmt.Errorf("encode google drive config: %w", err)
	}
	if _, err := s.db.Exec(`
		INSERT INTO app_settings (key, value, updated_at)
		VALUES ('google_drive', ?, ?)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			updated_at = excluded.updated_at
	`, string(payload), now.Format(time.RFC3339)); err != nil {
		return GoogleDriveConfig{}, fmt.Errorf("save google drive config: %w", err)
	}
	return cfg, nil
}

func (s *Store) ListRooms() ([]RoomRecord, error) {
	rows, err := s.db.Query(`
		SELECT room_id, floor, status, stay_mode, case_code, controller_online,
			remaining_minutes, last_payment_amount, last_payment_method, last_payment_ref, updated_at
		FROM rooms
		ORDER BY room_id
	`)
	if err != nil {
		return nil, fmt.Errorf("list rooms: %w", err)
	}
	defer rows.Close()

	var items []RoomRecord
	for rows.Next() {
		var item RoomRecord
		var updatedAt string
		var controllerOnline int
		if err := rows.Scan(
			&item.RoomID,
			&item.Floor,
			&item.Status,
			&item.StayMode,
			&item.CaseCode,
			&controllerOnline,
			&item.RemainingMinutes,
			&item.LastPaymentAmount,
			&item.LastPaymentMethod,
			&item.LastPaymentRef,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan room: %w", err)
		}
		item.ControllerOnline = controllerOnline == 1
		item.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse room updated_at: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) ListRecentAudit(roomID string, limit int) ([]AuditRecord, error) {
	query := `
		SELECT event_id, event_time, room_id, action, actor, result, stay_mode, case_code, detail
		FROM audit_events
	`
	var args []any
	if roomID != "" {
		query += ` WHERE room_id = ?`
		args = append(args, roomID)
	}
	query += ` ORDER BY event_time DESC, event_id DESC`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list audit: %w", err)
	}
	defer rows.Close()

	var items []AuditRecord
	for rows.Next() {
		var item AuditRecord
		var timestamp string
		if err := rows.Scan(
			&item.ID,
			&timestamp,
			&item.RoomID,
			&item.Action,
			&item.Actor,
			&item.Result,
			&item.StayMode,
			&item.CaseCode,
			&item.Detail,
		); err != nil {
			return nil, fmt.Errorf("scan audit: %w", err)
		}
		item.Timestamp, err = time.Parse(time.RFC3339, timestamp)
		if err != nil {
			return nil, fmt.Errorf("parse audit timestamp: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) RecordCheckIn(ctx context.Context, params CheckInParams) error {
	return s.withTx(ctx, func(tx *sql.Tx) error {
		if err := updateRoomState(tx, roomUpdate{
			RoomID:            params.RoomID,
			Status:            "occupied",
			StayMode:          params.StayMode,
			CaseCode:          params.CaseCode,
			ControllerOnline:  true,
			RemainingMinutes:  params.RemainingMinutes,
			LastPaymentAmount: params.PaymentAmount,
			LastPaymentMethod: choosePaymentMethod(params.PaymentAmount, "check_in"),
			LastPaymentRef:    params.PaymentRef,
			UpdatedAt:         params.Timestamp,
		}); err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO stays (
				stay_id, request_id, room_id, stay_mode, case_code, status,
				guest_in_at, allocated_minutes, payment_amount, payment_ref, last_updated_at
			) VALUES (?, ?, ?, ?, ?, 'in_use', ?, ?, ?, ?, ?)
		`,
			params.RequestID,
			params.RequestID,
			params.RoomID,
			params.StayMode,
			params.CaseCode,
			params.Timestamp.Format(time.RFC3339),
			params.RemainingMinutes,
			params.PaymentAmount,
			params.PaymentRef,
			params.Timestamp.Format(time.RFC3339),
		); err != nil {
			return fmt.Errorf("insert stay checkin: %w", err)
		}

		if params.PaymentAmount > 0 {
			if err := insertPayment(ctx, tx, PaymentRecord{
				PaymentID: params.RequestID + "-checkin",
				RequestID: params.RequestID,
				StayID:    params.RequestID,
				RoomID:    params.RoomID,
				Amount:    params.PaymentAmount,
				Method:    "check_in",
				Reference: params.PaymentRef,
				PaidAt:    params.Timestamp,
			}); err != nil {
				return err
			}
		}

		return insertAudit(ctx, tx, AuditRecord{
			ID:        params.RequestID,
			Timestamp: params.Timestamp,
			RoomID:    params.RoomID,
			Action:    "check_in",
			Actor:     params.Actor,
			Result:    "accepted",
			StayMode:  params.StayMode,
			CaseCode:  params.CaseCode,
			Detail:    params.PaymentRef,
		})
	})
}

func (s *Store) RecordCheckOut(ctx context.Context, params CheckOutParams) error {
	return s.withTx(ctx, func(tx *sql.Tx) error {
		stayID, guestInAt, err := findActiveStay(ctx, tx, params.RoomID)
		if err != nil {
			return err
		}
		if stayID != "" {
			usedMinutes := diffMinutes(guestInAt, params.Timestamp)
			if _, err := tx.ExecContext(ctx, `
				UPDATE stays
				SET status = 'checked_out', guest_out_at = ?, used_minutes = ?, last_updated_at = ?
				WHERE stay_id = ?
			`,
				params.Timestamp.Format(time.RFC3339),
				usedMinutes,
				params.Timestamp.Format(time.RFC3339),
				stayID,
			); err != nil {
				return fmt.Errorf("update stay checkout: %w", err)
			}
		}

		if err := updateRoomState(tx, roomUpdate{
			RoomID:           params.RoomID,
			Status:           "vacant",
			StayMode:         "",
			CaseCode:         "",
			ControllerOnline: true,
			RemainingMinutes: 0,
			UpdatedAt:        params.Timestamp,
		}); err != nil {
			return err
		}

		return insertAudit(ctx, tx, AuditRecord{
			ID:        params.RequestID,
			Timestamp: params.Timestamp,
			RoomID:    params.RoomID,
			Action:    "check_out",
			Actor:     params.Actor,
			Result:    "accepted",
		})
	})
}

func (s *Store) RecordExtend(ctx context.Context, params ExtendParams) error {
	return s.withTx(ctx, func(tx *sql.Tx) error {
		stayID, _, err := findActiveStay(ctx, tx, params.RoomID)
		if err != nil {
			return err
		}
		if stayID != "" {
			if _, err := tx.ExecContext(ctx, `
				UPDATE stays
				SET stay_mode = ?, case_code = ?, allocated_minutes = ?, payment_amount = payment_amount + ?,
					payment_ref = ?, last_updated_at = ?, remark = TRIM(COALESCE(remark, '') || CASE WHEN COALESCE(remark, '') = '' THEN '' ELSE '; ' END || ?)
				WHERE stay_id = ?
			`,
				params.StayMode,
				params.CaseCode,
				params.RemainingMinutes,
				params.PaymentAmount,
				params.PaymentRef,
				params.Timestamp.Format(time.RFC3339),
				fmt.Sprintf("extend_minutes=%d", params.ExtendMinutes),
				stayID,
			); err != nil {
				return fmt.Errorf("update stay extend: %w", err)
			}
		}

		if err := updateRoomState(tx, roomUpdate{
			RoomID:            params.RoomID,
			Status:            "occupied",
			StayMode:          params.StayMode,
			CaseCode:          params.CaseCode,
			ControllerOnline:  true,
			RemainingMinutes:  params.RemainingMinutes,
			LastPaymentAmount: params.PaymentAmount,
			LastPaymentMethod: choosePaymentMethod(params.PaymentAmount, "extend"),
			LastPaymentRef:    params.PaymentRef,
			UpdatedAt:         params.Timestamp,
		}); err != nil {
			return err
		}

		if params.PaymentAmount > 0 {
			if err := insertPayment(ctx, tx, PaymentRecord{
				PaymentID: params.RequestID + "-extend",
				RequestID: params.RequestID,
				StayID:    stayID,
				RoomID:    params.RoomID,
				Amount:    params.PaymentAmount,
				Method:    "extend",
				Reference: params.PaymentRef,
				PaidAt:    params.Timestamp,
			}); err != nil {
				return err
			}
		}

		return insertAudit(ctx, tx, AuditRecord{
			ID:        params.RequestID,
			Timestamp: params.Timestamp,
			RoomID:    params.RoomID,
			Action:    "extend_stay",
			Actor:     params.Actor,
			Result:    "accepted",
			StayMode:  params.StayMode,
			CaseCode:  params.CaseCode,
			Detail:    params.PaymentRef,
		})
	})
}

func (s *Store) RecordPayment(ctx context.Context, params PaymentParams) error {
	return s.withTx(ctx, func(tx *sql.Tx) error {
		stayID, _, err := findActiveStay(ctx, tx, params.RoomID)
		if err != nil {
			return err
		}

		if err := updateRoomPayment(tx, params.RoomID, params.Amount, params.Method, params.Reference, params.Timestamp); err != nil {
			return err
		}

		if err := insertPayment(ctx, tx, PaymentRecord{
			PaymentID: params.RequestID,
			RequestID: params.RequestID,
			StayID:    stayID,
			RoomID:    params.RoomID,
			Amount:    params.Amount,
			Method:    params.Method,
			Reference: params.Reference,
			PaidAt:    params.Timestamp,
		}); err != nil {
			return err
		}

		roomMeta, err := roomMetaByID(ctx, tx, params.RoomID)
		if err != nil {
			return err
		}

		return insertAudit(ctx, tx, AuditRecord{
			ID:        params.RequestID,
			Timestamp: params.Timestamp,
			RoomID:    params.RoomID,
			Action:    "receive_payment",
			Actor:     params.Actor,
			Result:    "accepted",
			StayMode:  roomMeta.StayMode,
			CaseCode:  roomMeta.CaseCode,
			Detail:    params.Reference,
		})
	})
}

func (s *Store) RecordCommand(ctx context.Context, params CommandParams) error {
	return s.withTx(ctx, func(tx *sql.Tx) error {
		switch params.Action {
		case "start_cleaning":
			if err := updateRoomState(tx, roomUpdate{
				RoomID:           params.RoomID,
				Status:           "cleaning",
				ControllerOnline: true,
				TouchOnly:        true,
				UpdatedAt:        params.Timestamp,
			}); err != nil {
				return err
			}
		case "finish_cleaning":
			if err := updateRoomState(tx, roomUpdate{
				RoomID:           params.RoomID,
				Status:           "vacant",
				StayMode:         "",
				CaseCode:         "",
				ControllerOnline: true,
				RemainingMinutes: 0,
				UpdatedAt:        params.Timestamp,
			}); err != nil {
				return err
			}
			if err := closeCleaningForLatestStay(ctx, tx, params.RoomID, params.Timestamp); err != nil {
				return err
			}
		}

		roomMeta, err := roomMetaByID(ctx, tx, params.RoomID)
		if err != nil {
			return err
		}

		return insertAudit(ctx, tx, AuditRecord{
			ID:        params.RequestID,
			Timestamp: params.Timestamp,
			RoomID:    params.RoomID,
			Action:    params.Action,
			Actor:     params.Actor,
			Result:    "accepted",
			StayMode:  roomMeta.StayMode,
			CaseCode:  roomMeta.CaseCode,
			Detail:    params.Detail,
		})
	})
}

func (s *Store) DailyWorkbookData(day time.Time, loc *time.Location) (DailyWorkbookData, error) {
	rooms, err := s.ListRooms()
	if err != nil {
		return DailyWorkbookData{}, err
	}
	stays, err := s.listStaysByDate(day, loc)
	if err != nil {
		return DailyWorkbookData{}, err
	}
	payments, err := s.listPaymentsByDate(day, loc)
	if err != nil {
		return DailyWorkbookData{}, err
	}
	audits, err := s.listAuditByDate(day, loc)
	if err != nil {
		return DailyWorkbookData{}, err
	}
	return DailyWorkbookData{
		Date:     day.In(loc).Format("2006-01-02"),
		Rooms:    rooms,
		Stays:    stays,
		Payments: payments,
		Audits:   audits,
	}, nil
}

func (s *Store) MonthlyWorkbookData(month time.Time, loc *time.Location) (MonthlyWorkbookData, error) {
	rooms, err := s.ListRooms()
	if err != nil {
		return MonthlyWorkbookData{}, err
	}
	stays, err := s.listStaysByMonth(month, loc)
	if err != nil {
		return MonthlyWorkbookData{}, err
	}
	payments, err := s.listPaymentsByMonth(month, loc)
	if err != nil {
		return MonthlyWorkbookData{}, err
	}
	audits, err := s.listAuditByMonth(month, loc)
	if err != nil {
		return MonthlyWorkbookData{}, err
	}
	return MonthlyWorkbookData{
		MonthKey: month.In(loc).Format("2006-01"),
		Rooms:    rooms,
		Stays:    stays,
		Payments: payments,
		Audits:   audits,
	}, nil
}

func (s *Store) CandidateDailyExportDates(now time.Time, loc *time.Location) ([]time.Time, error) {
	timestamps := []string{}
	sources := []string{
		`SELECT guest_in_at FROM stays WHERE guest_in_at IS NOT NULL`,
		`SELECT guest_out_at FROM stays WHERE guest_out_at IS NOT NULL`,
		`SELECT clean_up_at FROM stays WHERE clean_up_at IS NOT NULL`,
		`SELECT paid_at FROM payments`,
		`SELECT event_time FROM audit_events`,
	}
	for _, query := range sources {
		rows, err := s.db.Query(query)
		if err != nil {
			return nil, fmt.Errorf("query export candidates: %w", err)
		}
		for rows.Next() {
			var raw string
			if err := rows.Scan(&raw); err != nil {
				rows.Close()
				return nil, fmt.Errorf("scan export candidate: %w", err)
			}
			if strings.TrimSpace(raw) != "" {
				timestamps = append(timestamps, raw)
			}
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, fmt.Errorf("iterate export candidates: %w", err)
		}
		rows.Close()
	}

	successful, err := s.successfulReportPeriods("daily")
	if err != nil {
		return nil, err
	}

	cutoff := now.In(loc).Format("2006-01-02")
	seen := make(map[string]time.Time)
	for _, raw := range timestamps {
		ts, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			continue
		}
		dateKey := ts.In(loc).Format("2006-01-02")
		if dateKey >= cutoff {
			continue
		}
		if _, ok := successful[dateKey]; ok {
			continue
		}
		if _, ok := seen[dateKey]; !ok {
			seen[dateKey] = time.Date(ts.In(loc).Year(), ts.In(loc).Month(), ts.In(loc).Day(), 0, 0, 0, 0, loc)
		}
	}

	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make([]time.Time, 0, len(keys))
	for _, key := range keys {
		result = append(result, seen[key])
	}
	return result, nil
}

func (s *Store) MarkExportRun(day time.Time, status, localPath, driveFileID, lastError string, uploadedAt *time.Time) error {
	if err := s.MarkReportRun("daily", day.Format("2006-01-02"), status, localPath, driveFileID, lastError, uploadedAt); err != nil {
		return err
	}
	_, err := s.db.Exec(`
		INSERT INTO export_runs (export_date, status, local_path, drive_file_id, last_error, uploaded_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(export_date) DO UPDATE SET
			status = excluded.status,
			local_path = excluded.local_path,
			drive_file_id = excluded.drive_file_id,
			last_error = excluded.last_error,
			uploaded_at = excluded.uploaded_at,
			updated_at = excluded.updated_at
	`,
		day.Format("2006-01-02"),
		status,
		localPath,
		driveFileID,
		lastError,
		formatNullableTime(uploadedAt),
		time.Now().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("mark export run: %w", err)
	}
	return nil
}

func (s *Store) MarkReportRun(reportKind, periodKey, status, localPath, driveFileID, lastError string, uploadedAt *time.Time) error {
	_, err := s.db.Exec(`
		INSERT INTO report_runs (report_kind, period_key, status, local_path, drive_file_id, last_error, uploaded_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(report_kind, period_key) DO UPDATE SET
			status = excluded.status,
			local_path = excluded.local_path,
			drive_file_id = excluded.drive_file_id,
			last_error = excluded.last_error,
			uploaded_at = excluded.uploaded_at,
			updated_at = excluded.updated_at
	`,
		reportKind,
		periodKey,
		status,
		localPath,
		driveFileID,
		lastError,
		formatNullableTime(uploadedAt),
		time.Now().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("mark report run: %w", err)
	}
	return nil
}

func (s *Store) successfulExportDates() (map[string]struct{}, error) {
	return s.successfulReportPeriods("daily")
}

func (s *Store) successfulReportPeriods(reportKind string) (map[string]struct{}, error) {
	rows, err := s.db.Query(`SELECT period_key FROM report_runs WHERE report_kind = ? AND status = 'uploaded'`, reportKind)
	if err != nil {
		return nil, fmt.Errorf("list successful report periods: %w", err)
	}
	defer rows.Close()

	result := make(map[string]struct{})
	for rows.Next() {
		var periodKey string
		if err := rows.Scan(&periodKey); err != nil {
			return nil, fmt.Errorf("scan successful report period: %w", err)
		}
		result[periodKey] = struct{}{}
	}
	return result, rows.Err()
}

func (s *Store) listStaysByDate(day time.Time, loc *time.Location) ([]StayRecord, error) {
	rows, err := s.db.Query(`
		SELECT stay_id, request_id, room_id, stay_mode, case_code, status, guest_in_at, guest_out_at, clean_up_at,
			allocated_minutes, used_minutes, payment_amount, payment_ref, remark, last_updated_at
		FROM stays
		ORDER BY room_id, guest_in_at
	`)
	if err != nil {
		return nil, fmt.Errorf("list stays for export: %w", err)
	}
	defer rows.Close()

	var items []StayRecord
	for rows.Next() {
		var item StayRecord
		var guestInAt, guestOutAt, cleanUpAt sql.NullString
		var lastUpdatedAt string
		if err := rows.Scan(
			&item.StayID,
			&item.RequestID,
			&item.RoomID,
			&item.StayMode,
			&item.CaseCode,
			&item.Status,
			&guestInAt,
			&guestOutAt,
			&cleanUpAt,
			&item.AllocatedMinutes,
			&item.UsedMinutes,
			&item.PaymentAmount,
			&item.PaymentRef,
			&item.Remark,
			&lastUpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan stay: %w", err)
		}
		item.GuestInAt = parseNullableTime(guestInAt)
		item.GuestOutAt = parseNullableTime(guestOutAt)
		item.CleanUpAt = parseNullableTime(cleanUpAt)
		item.LastUpdatedAt, err = time.Parse(time.RFC3339, lastUpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse stay last_updated_at: %w", err)
		}
		if stayMatchesDate(item, day, loc) {
			items = append(items, item)
		}
	}
	return items, rows.Err()
}

func (s *Store) listPaymentsByDate(day time.Time, loc *time.Location) ([]PaymentRecord, error) {
	rows, err := s.db.Query(`
		SELECT payment_id, request_id, stay_id, room_id, amount, method, reference, paid_at
		FROM payments
		ORDER BY paid_at, payment_id
	`)
	if err != nil {
		return nil, fmt.Errorf("list payments for export: %w", err)
	}
	defer rows.Close()

	var items []PaymentRecord
	for rows.Next() {
		var item PaymentRecord
		var paidAt string
		if err := rows.Scan(
			&item.PaymentID,
			&item.RequestID,
			&item.StayID,
			&item.RoomID,
			&item.Amount,
			&item.Method,
			&item.Reference,
			&paidAt,
		); err != nil {
			return nil, fmt.Errorf("scan payment: %w", err)
		}
		item.PaidAt, err = time.Parse(time.RFC3339, paidAt)
		if err != nil {
			return nil, fmt.Errorf("parse payment paid_at: %w", err)
		}
		if sameLocalDate(item.PaidAt, day, loc) {
			items = append(items, item)
		}
	}
	return items, rows.Err()
}

func (s *Store) listAuditByDate(day time.Time, loc *time.Location) ([]AuditRecord, error) {
	rows, err := s.db.Query(`
		SELECT event_id, event_time, room_id, action, actor, result, stay_mode, case_code, detail
		FROM audit_events
		ORDER BY event_time, event_id
	`)
	if err != nil {
		return nil, fmt.Errorf("list audit for export: %w", err)
	}
	defer rows.Close()

	var items []AuditRecord
	for rows.Next() {
		var item AuditRecord
		var timestamp string
		if err := rows.Scan(
			&item.ID,
			&timestamp,
			&item.RoomID,
			&item.Action,
			&item.Actor,
			&item.Result,
			&item.StayMode,
			&item.CaseCode,
			&item.Detail,
		); err != nil {
			return nil, fmt.Errorf("scan export audit: %w", err)
		}
		item.Timestamp, err = time.Parse(time.RFC3339, timestamp)
		if err != nil {
			return nil, fmt.Errorf("parse export audit timestamp: %w", err)
		}
		if sameLocalDate(item.Timestamp, day, loc) {
			items = append(items, item)
		}
	}
	return items, rows.Err()
}

func (s *Store) CandidateMonthlyExportMonths(now time.Time, loc *time.Location) ([]time.Time, error) {
	days, err := s.CandidateDailyExportDates(now, loc)
	if err != nil {
		return nil, err
	}
	successful, err := s.successfulReportPeriods("monthly")
	if err != nil {
		return nil, err
	}
	currentMonth := now.In(loc).Format("2006-01")
	seen := make(map[string]time.Time)
	for _, day := range days {
		monthKey := day.In(loc).Format("2006-01")
		if monthKey >= currentMonth {
			continue
		}
		if _, ok := successful[monthKey]; ok {
			continue
		}
		if _, ok := seen[monthKey]; !ok {
			seen[monthKey] = time.Date(day.In(loc).Year(), day.In(loc).Month(), 1, 0, 0, 0, 0, loc)
		}
	}
	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]time.Time, 0, len(keys))
	for _, key := range keys {
		result = append(result, seen[key])
	}
	return result, nil
}

func (s *Store) listStaysByMonth(month time.Time, loc *time.Location) ([]StayRecord, error) {
	return s.listStaysByDateRange(monthStart(month, loc), monthEnd(month, loc), loc)
}

func (s *Store) listPaymentsByMonth(month time.Time, loc *time.Location) ([]PaymentRecord, error) {
	items, err := s.listPaymentsAll()
	if err != nil {
		return nil, err
	}
	var result []PaymentRecord
	for _, item := range items {
		if sameLocalMonth(item.PaidAt, month, loc) {
			result = append(result, item)
		}
	}
	return result, nil
}

func (s *Store) listAuditByMonth(month time.Time, loc *time.Location) ([]AuditRecord, error) {
	items, err := s.listAuditAll()
	if err != nil {
		return nil, err
	}
	var result []AuditRecord
	for _, item := range items {
		if sameLocalMonth(item.Timestamp, month, loc) {
			result = append(result, item)
		}
	}
	return result, nil
}

func (s *Store) listStaysByDateRange(start, end time.Time, loc *time.Location) ([]StayRecord, error) {
	rows, err := s.db.Query(`
		SELECT stay_id, request_id, room_id, stay_mode, case_code, status, guest_in_at, guest_out_at, clean_up_at,
			allocated_minutes, used_minutes, payment_amount, payment_ref, remark, last_updated_at
		FROM stays
		ORDER BY room_id, guest_in_at
	`)
	if err != nil {
		return nil, fmt.Errorf("list stays in range: %w", err)
	}
	defer rows.Close()

	var items []StayRecord
	for rows.Next() {
		var item StayRecord
		var guestInAt, guestOutAt, cleanUpAt sql.NullString
		var lastUpdatedAt string
		if err := rows.Scan(
			&item.StayID, &item.RequestID, &item.RoomID, &item.StayMode, &item.CaseCode, &item.Status,
			&guestInAt, &guestOutAt, &cleanUpAt, &item.AllocatedMinutes, &item.UsedMinutes,
			&item.PaymentAmount, &item.PaymentRef, &item.Remark, &lastUpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan stay range: %w", err)
		}
		item.GuestInAt = parseNullableTime(guestInAt)
		item.GuestOutAt = parseNullableTime(guestOutAt)
		item.CleanUpAt = parseNullableTime(cleanUpAt)
		item.LastUpdatedAt, err = time.Parse(time.RFC3339, lastUpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse stay range last_updated_at: %w", err)
		}
		if stayMatchesDateRange(item, start, end, loc) {
			items = append(items, item)
		}
	}
	return items, rows.Err()
}

func (s *Store) listPaymentsAll() ([]PaymentRecord, error) {
	rows, err := s.db.Query(`
		SELECT payment_id, request_id, stay_id, room_id, amount, method, reference, paid_at
		FROM payments
		ORDER BY paid_at, payment_id
	`)
	if err != nil {
		return nil, fmt.Errorf("list payments all: %w", err)
	}
	defer rows.Close()
	var items []PaymentRecord
	for rows.Next() {
		var item PaymentRecord
		var paidAt string
		if err := rows.Scan(&item.PaymentID, &item.RequestID, &item.StayID, &item.RoomID, &item.Amount, &item.Method, &item.Reference, &paidAt); err != nil {
			return nil, fmt.Errorf("scan payment all: %w", err)
		}
		item.PaidAt, err = time.Parse(time.RFC3339, paidAt)
		if err != nil {
			return nil, fmt.Errorf("parse payment all paid_at: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) listAuditAll() ([]AuditRecord, error) {
	rows, err := s.db.Query(`
		SELECT event_id, event_time, room_id, action, actor, result, stay_mode, case_code, detail
		FROM audit_events
		ORDER BY event_time, event_id
	`)
	if err != nil {
		return nil, fmt.Errorf("list audit all: %w", err)
	}
	defer rows.Close()
	var items []AuditRecord
	for rows.Next() {
		var item AuditRecord
		var timestamp string
		if err := rows.Scan(&item.ID, &timestamp, &item.RoomID, &item.Action, &item.Actor, &item.Result, &item.StayMode, &item.CaseCode, &item.Detail); err != nil {
			return nil, fmt.Errorf("scan audit all: %w", err)
		}
		item.Timestamp, err = time.Parse(time.RFC3339, timestamp)
		if err != nil {
			return nil, fmt.Errorf("parse audit all timestamp: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) withTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin sqlite tx: %w", err)
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit sqlite tx: %w", err)
	}
	return nil
}

type roomUpdate struct {
	RoomID            string
	Status            string
	StayMode          string
	CaseCode          string
	ControllerOnline  bool
	RemainingMinutes  int
	LastPaymentAmount float64
	LastPaymentMethod string
	LastPaymentRef    string
	TouchOnly         bool
	UpdatedAt         time.Time
}

func updateRoomState(tx *sql.Tx, params roomUpdate) error {
	controllerOnline := 0
	if params.ControllerOnline {
		controllerOnline = 1
	}
	query := `
		UPDATE rooms
		SET status = ?, stay_mode = ?, case_code = ?, controller_online = ?, remaining_minutes = ?,
			last_payment_amount = CASE WHEN ? != 0 THEN ? ELSE last_payment_amount END,
			last_payment_method = CASE WHEN ? != '' THEN ? ELSE last_payment_method END,
			last_payment_ref = CASE WHEN ? != '' THEN ? ELSE last_payment_ref END,
			updated_at = ?
		WHERE room_id = ?
	`
	args := []any{
		params.Status,
		params.StayMode,
		params.CaseCode,
		controllerOnline,
		params.RemainingMinutes,
		params.LastPaymentAmount,
		params.LastPaymentAmount,
		params.LastPaymentMethod,
		params.LastPaymentMethod,
		params.LastPaymentRef,
		params.LastPaymentRef,
		params.UpdatedAt.Format(time.RFC3339),
		params.RoomID,
	}
	if params.TouchOnly {
		query = `
			UPDATE rooms
			SET status = ?, controller_online = ?, updated_at = ?
			WHERE room_id = ?
		`
		args = []any{
			params.Status,
			controllerOnline,
			params.UpdatedAt.Format(time.RFC3339),
			params.RoomID,
		}
	}
	if _, err := tx.Exec(query, args...); err != nil {
		return fmt.Errorf("update room state: %w", err)
	}
	return nil
}

func updateRoomPayment(tx *sql.Tx, roomID string, amount float64, method, ref string, updatedAt time.Time) error {
	if _, err := tx.Exec(`
		UPDATE rooms
		SET last_payment_amount = ?, last_payment_method = ?, last_payment_ref = ?, updated_at = ?
		WHERE room_id = ?
	`,
		amount,
		method,
		ref,
		updatedAt.Format(time.RFC3339),
		roomID,
	); err != nil {
		return fmt.Errorf("update room payment: %w", err)
	}
	return nil
}

func insertPayment(ctx context.Context, tx *sql.Tx, payment PaymentRecord) error {
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO payments (payment_id, request_id, stay_id, room_id, amount, method, reference, paid_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		payment.PaymentID,
		payment.RequestID,
		payment.StayID,
		payment.RoomID,
		payment.Amount,
		payment.Method,
		payment.Reference,
		payment.PaidAt.Format(time.RFC3339),
	); err != nil {
		return fmt.Errorf("insert payment: %w", err)
	}
	return nil
}

func insertAudit(ctx context.Context, tx *sql.Tx, audit AuditRecord) error {
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO audit_events (event_id, room_id, action, actor, result, stay_mode, case_code, detail, event_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		audit.ID,
		audit.RoomID,
		audit.Action,
		audit.Actor,
		audit.Result,
		audit.StayMode,
		audit.CaseCode,
		audit.Detail,
		audit.Timestamp.Format(time.RFC3339),
	); err != nil {
		return fmt.Errorf("insert audit: %w", err)
	}
	return nil
}

func findActiveStay(ctx context.Context, tx *sql.Tx, roomID string) (string, time.Time, error) {
	var stayID string
	var guestInRaw sql.NullString
	err := tx.QueryRowContext(ctx, `
		SELECT stay_id, guest_in_at
		FROM stays
		WHERE room_id = ? AND status = 'in_use'
		ORDER BY guest_in_at DESC, stay_id DESC
		LIMIT 1
	`, roomID).Scan(&stayID, &guestInRaw)
	if err == sql.ErrNoRows {
		return "", time.Time{}, nil
	}
	if err != nil {
		return "", time.Time{}, fmt.Errorf("find active stay: %w", err)
	}
	if !guestInRaw.Valid || strings.TrimSpace(guestInRaw.String) == "" {
		return stayID, time.Time{}, nil
	}
	guestInAt, err := time.Parse(time.RFC3339, guestInRaw.String)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("parse active stay guest_in_at: %w", err)
	}
	return stayID, guestInAt, nil
}

func closeCleaningForLatestStay(ctx context.Context, tx *sql.Tx, roomID string, ts time.Time) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE stays
		SET clean_up_at = COALESCE(clean_up_at, ?), last_updated_at = ?
		WHERE stay_id = (
			SELECT stay_id
			FROM stays
			WHERE room_id = ? AND guest_out_at IS NOT NULL
			ORDER BY guest_out_at DESC, stay_id DESC
			LIMIT 1
		)
	`,
		ts.Format(time.RFC3339),
		ts.Format(time.RFC3339),
		roomID,
	)
	if err != nil {
		return fmt.Errorf("mark clean up stay: %w", err)
	}
	return nil
}

type roomMeta struct {
	StayMode string
	CaseCode string
}

func roomMetaByID(ctx context.Context, tx *sql.Tx, roomID string) (roomMeta, error) {
	var meta roomMeta
	if err := tx.QueryRowContext(ctx, `
		SELECT stay_mode, case_code
		FROM rooms
		WHERE room_id = ?
	`, roomID).Scan(&meta.StayMode, &meta.CaseCode); err != nil {
		return roomMeta{}, fmt.Errorf("load room meta: %w", err)
	}
	return meta, nil
}

func choosePaymentMethod(amount float64, fallback string) string {
	if amount <= 0 {
		return ""
	}
	return fallback
}

func parseNullableTime(value sql.NullString) *time.Time {
	if !value.Valid || strings.TrimSpace(value.String) == "" {
		return nil
	}
	ts, err := time.Parse(time.RFC3339, value.String)
	if err != nil {
		return nil
	}
	return &ts
}

func formatNullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.Format(time.RFC3339)
}

func sameLocalDate(ts, day time.Time, loc *time.Location) bool {
	return ts.In(loc).Format("2006-01-02") == day.In(loc).Format("2006-01-02")
}

func sameLocalMonth(ts, month time.Time, loc *time.Location) bool {
	return ts.In(loc).Format("2006-01") == month.In(loc).Format("2006-01")
}

func stayMatchesDate(stay StayRecord, day time.Time, loc *time.Location) bool {
	candidates := []*time.Time{stay.GuestInAt, stay.GuestOutAt, stay.CleanUpAt, &stay.LastUpdatedAt}
	for _, ts := range candidates {
		if ts != nil && sameLocalDate(*ts, day, loc) {
			return true
		}
	}
	return false
}

func stayMatchesDateRange(stay StayRecord, start, end time.Time, loc *time.Location) bool {
	candidates := []*time.Time{stay.GuestInAt, stay.GuestOutAt, stay.CleanUpAt, &stay.LastUpdatedAt}
	for _, ts := range candidates {
		if ts == nil {
			continue
		}
		local := ts.In(loc)
		if !local.Before(start.In(loc)) && !local.After(end.In(loc)) {
			return true
		}
	}
	return false
}

func diffMinutes(start, end time.Time) int {
	if start.IsZero() || end.Before(start) {
		return 0
	}
	return int(end.Sub(start).Minutes())
}

func floorFromRoomID(roomID string) int {
	if len(roomID) == 0 {
		return 0
	}
	return int(roomID[0] - '0')
}

func monthStart(month time.Time, loc *time.Location) time.Time {
	local := month.In(loc)
	return time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, loc)
}

func monthEnd(month time.Time, loc *time.Location) time.Time {
	return monthStart(month, loc).AddDate(0, 1, 0).Add(-time.Nanosecond)
}
