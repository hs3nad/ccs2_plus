package transport

import (
	"context"
	"fmt"
	"time"

	"ccs2plus/go-backend/internal/legacy"
	"ccs2plus/go-backend/internal/roommap"
)

type CommandIntent struct {
	RequestID    string
	RoomID       string
	Action       string
	TargetState  legacy.SemanticState
	StayMode     string
	CaseCode     string
}

type CommandResult struct {
	RequestID       string
	RoomID          string
	Action          string
	TargetState     string
	StayMode        string
	CaseCode        string
	SlaveFrame      []byte
	DisplayFrame    []byte
	SentAt          time.Time
}

type Writer interface {
	WriteFrame(ctx context.Context, frame []byte) error
}

type Service struct {
	repo   roommap.Repository
	writer Writer
	cache  map[byte]legacy.SlavePayload
}

func NewService(repo roommap.Repository, writer Writer) *Service {
	return &Service{
		repo:   repo,
		writer: writer,
		cache:  make(map[byte]legacy.SlavePayload),
	}
}

func (s *Service) ApplyIntent(ctx context.Context, intent CommandIntent) (CommandResult, error) {
	mapping, ok := s.repo.GetByRoomID(intent.RoomID)
	if !ok {
		return CommandResult{}, fmt.Errorf("room mapping not found: %s", intent.RoomID)
	}

	roomStatus, displayMode, ok := legacy.SemanticToLegacy(intent.TargetState)
	if !ok {
		return CommandResult{}, fmt.Errorf("unsupported target state: %s", intent.TargetState)
	}

	payload, ok := s.cache[mapping.Address]
	if !ok {
		payload = legacy.EmptyPayload()
	}

	if err := payload.SetPortStatus(mapping.PortIndex, roomStatus); err != nil {
		return CommandResult{}, err
	}
	s.cache[mapping.Address] = payload

	slaveFrame := legacy.BuildSlaveFrame(mapping.Address, payload)
	displayFrame := legacy.BuildDisplayFrame(mapping.CardIndex, mapping.PortIndex, displayMode)

	if s.writer != nil {
		if err := s.writer.WriteFrame(ctx, slaveFrame); err != nil {
			return CommandResult{}, err
		}
		if err := s.writer.WriteFrame(ctx, displayFrame); err != nil {
			return CommandResult{}, err
		}
	}

	return CommandResult{
		RequestID:    intent.RequestID,
		RoomID:       intent.RoomID,
		Action:       intent.Action,
		TargetState:  string(intent.TargetState),
		StayMode:     intent.StayMode,
		CaseCode:     intent.CaseCode,
		SlaveFrame:   slaveFrame,
		DisplayFrame: displayFrame,
		SentAt:       time.Now(),
	}, nil
}
