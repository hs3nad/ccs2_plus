package transport

import (
	"bytes"
	"context"
	"testing"

	"ccs2plus/go-backend/internal/legacy"
	"ccs2plus/go-backend/internal/roommap"
)

func TestProtocolDocExamplesAmpV2D(t *testing.T) {
	tests := []struct {
		name        string
		intent      CommandIntent
		prepare     []CommandIntent
		wantSlave   []byte
		wantDisplay []byte
	}{
		{
			name: "docs_check_in",
			intent: CommandIntent{
				RequestID:   "doc-in",
				RoomID:      "1205",
				Action:      "check_in",
				TargetState: legacy.SemanticOccupied,
				StayMode:    "overnight",
				CaseCode:    "overnight",
			},
			wantSlave:   []byte{0x3a, 0x03, 0x20, 0x00, 0xa3},
			wantDisplay: []byte{0x3a, 0xfe, 0x35, 0x03},
		},
		{
			name: "docs_check_out",
			prepare: []CommandIntent{
				{
					RequestID:   "doc-prep-out",
					RoomID:      "1205",
					Action:      "check_in",
					TargetState: legacy.SemanticOccupied,
					StayMode:    "overnight",
					CaseCode:    "overnight",
				},
			},
			intent: CommandIntent{
				RequestID:   "doc-out",
				RoomID:      "1205",
				Action:      "check_out",
				TargetState: legacy.SemanticVacant,
			},
			wantSlave:   []byte{0x3a, 0x03, 0x00, 0x00, 0xc3},
			wantDisplay: []byte{0x3a, 0xfe, 0x35, 0x00},
		},
		{
			name: "docs_start_cleaning",
			intent: CommandIntent{
				RequestID:   "doc-clean",
				RoomID:      "1205",
				Action:      "start_cleaning",
				TargetState: legacy.SemanticCleaning,
			},
			wantSlave:   []byte{0x3a, 0x03, 0x20, 0x00, 0xa3},
			wantDisplay: []byte{0x3a, 0xfe, 0x35, 0x03},
		},
		{
			name: "docs_overstay",
			intent: CommandIntent{
				RequestID:   "doc-ot",
				RoomID:      "1205",
				Action:      "mark_overstay",
				TargetState: legacy.SemanticOverstay,
				StayMode:    "temporary",
				CaseCode:    "temporary",
			},
			wantSlave:   []byte{0x3a, 0x03, 0x20, 0x00, 0xa3},
			wantDisplay: []byte{0x3a, 0xfe, 0x35, 0x03},
		},
		{
			name: "docs_extend_stay",
			intent: CommandIntent{
				RequestID:   "doc-extend",
				RoomID:      "1205",
				Action:      "extend_stay",
				TargetState: legacy.SemanticOccupied,
				StayMode:    "temporary",
				CaseCode:    "temporary",
			},
			wantSlave:   []byte{0x3a, 0x03, 0x20, 0x00, 0xa3},
			wantDisplay: []byte{0x3a, 0xfe, 0x35, 0x03},
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
				t.Fatalf("apply intent failed: %v", err)
			}

			if !bytes.Equal(result.SlaveFrame, tt.wantSlave) {
				t.Fatalf("doc example slave mismatch: got % X want % X", result.SlaveFrame, tt.wantSlave)
			}
			if !bytes.Equal(result.DisplayFrame, tt.wantDisplay) {
				t.Fatalf("doc example display mismatch: got % X want % X", result.DisplayFrame, tt.wantDisplay)
			}
		})
	}
}
