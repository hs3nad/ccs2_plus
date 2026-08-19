package config

import (
	"encoding/json"
	"fmt"
	"os"

	"ccs2plus/go-backend/internal/roommap"
)

type RoomMapFile struct {
	Site  SiteInfo               `json:"site"`
	Rooms []roommap.RoomMapping  `json:"rooms"`
}

type SiteInfo struct {
	Name        string `json:"name"`
	Controller  string `json:"controller"`
	Transport   string `json:"transport"`
	Topology    string `json:"topology"`
}

func LoadRoomMapFile(path string) (RoomMapFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return RoomMapFile{}, fmt.Errorf("read room map file: %w", err)
	}

	var cfg RoomMapFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		return RoomMapFile{}, fmt.Errorf("decode room map file: %w", err)
	}

	return cfg, nil
}
