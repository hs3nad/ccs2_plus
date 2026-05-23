package transport

import (
	"bytes"
	"context"
	"testing"

	"ccs2plus/go-backend/internal/legacy"
	"ccs2plus/go-backend/internal/roommap"
)

func TestApplyIntentAmpV2DFrames(t *testing.T) {
	tests := []struct {
		name         string
		intent       CommandIntent
		prepare      []CommandIntent
		wantSlave    []byte
		wantDisplay  []byte
		wantTarget   string
	}{
		{
			name: "check_in",
			intent: CommandIntent{
				RequestID:   "req-in",
				RoomID:      "1205",
				Action:      "check_in",
				TargetState: legacy.SemanticOccupied,
				StayMode:    "overnight",
				CaseCode:    "overnight",
			},
			wantSlave:   []byte{0x3a, 0x03, 0x20, 0x00, 0xa3},
			wantDisplay: []byte{0x3a, 0xfe, 0x35, 0x03},
			wantTarget:  "occupied",
		},
		{
			name: "check_out",
			prepare: []CommandIntent{
				{
					RequestID:   "prep-out",
					RoomID:      "1205",
					Action:      "check_in",
					TargetState: legacy.SemanticOccupied,
					StayMode:    "overnight",
					CaseCode:    "overnight",
				},
			},
			intent: CommandIntent{
				RequestID:   "req-out",
				RoomID:      "1205",
				Action:      "check_out",
				TargetState: legacy.SemanticVacant,
			},
			wantSlave:   []byte{0x3a, 0x03, 0x00, 0x00, 0xc3},
			wantDisplay: []byte{0x3a, 0xfe, 0x35, 0x00},
			wantTarget:  "vacant",
		},
		{
			name: "start_cleaning",
			intent: CommandIntent{
				RequestID:   "req-clean",
				RoomID:      "1205",
				Action:      "start_cleaning",
				TargetState: legacy.SemanticCleaning,
			},
			wantSlave:   []byte{0x3a, 0x03, 0x20, 0x00, 0xa3},
			wantDisplay: []byte{0x3a, 0xfe, 0x35, 0x03},
			wantTarget:  "cleaning",
		},
		{
			name: "overstay",
			intent: CommandIntent{
				RequestID:   "req-ot",
				RoomID:      "1205",
				Action:      "mark_overstay",
				TargetState: legacy.SemanticOverstay,
				StayMode:    "temporary",
				CaseCode:    "temporary",
			},
			wantSlave:   []byte{0x3a, 0x03, 0x20, 0x00, 0xa3},
			wantDisplay: []byte{0x3a, 0xfe, 0x35, 0x03},
			wantTarget:  "overstay",
		},
		{
			name: "extend_stay",
			intent: CommandIntent{
				RequestID:   "req-extend",
				RoomID:      "1205",
				Action:      "extend_stay",
				TargetState: legacy.SemanticOccupied,
				StayMode:    "temporary",
				CaseCode:    "temporary",
			},
			wantSlave:   []byte{0x3a, 0x03, 0x20, 0x00, 0xa3},
			wantDisplay: []byte{0x3a, 0xfe, 0x35, 0x03},
			wantTarget:  "occupied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := roommap.NewMemoryRepository([]roommap.RoomMapping{
				{
					RoomID:       "1205",
					ControllerID: "ctrl-1205",
					Address:      0x03,
					CardIndex:    3,
					PortIndex:    5,
				},
			})
			writer := &MemoryWriter{}
			svc := NewService(repo, writer)

			for _, prep := range tt.prepare {
				if _, err := svc.ApplyIntent(context.Background(), prep); err != nil {
					t.Fatalf("prepare intent failed: %v", err)
				}
				writer.Frames = nil
			}

			result, err := svc.ApplyIntent(context.Background(), tt.intent)
			if err != nil {
				t.Fatalf("apply intent: %v", err)
			}

			if !bytes.Equal(result.SlaveFrame, tt.wantSlave) {
				t.Fatalf("unexpected slave frame: got % X want % X", result.SlaveFrame, tt.wantSlave)
			}

			if !bytes.Equal(result.DisplayFrame, tt.wantDisplay) {
				t.Fatalf("unexpected display frame: got % X want % X", result.DisplayFrame, tt.wantDisplay)
			}

			if result.TargetState != tt.wantTarget {
				t.Fatalf("unexpected target state: got %s want %s", result.TargetState, tt.wantTarget)
			}

			if len(writer.Frames) != 2 {
				t.Fatalf("expected 2 written frames, got %d", len(writer.Frames))
			}
		})
	}
}
