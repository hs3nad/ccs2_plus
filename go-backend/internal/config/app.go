package config

import (
	"flag"
	"fmt"
)

type AppConfig struct {
	ListenAddr             string
	RoomMapPath            string
	SerialDevice           string
	SerialBaud             int
	WriteDelayMS           int
	DBPath                 string
	ExportDir              string
	ExportTimezone         string
	GoogleDriveCredentials string
	GoogleDriveFolderID    string
	DashboardDir           string
}

func LoadAppConfig() (AppConfig, error) {
	cfg := AppConfig{}

	flag.StringVar(&cfg.ListenAddr, "listen", ":8080", "HTTP listen address")
	flag.StringVar(&cfg.RoomMapPath, "roommap", "", "path to room mapping JSON file")
	flag.StringVar(&cfg.SerialDevice, "serial", "", "USB-RS485 serial device path, e.g. /dev/ttyUSB0")
	flag.IntVar(&cfg.SerialBaud, "baud", 9600, "serial baud rate")
	flag.IntVar(&cfg.WriteDelayMS, "write-delay-ms", 30, "inter-frame write delay in milliseconds")
	flag.StringVar(&cfg.DBPath, "db-path", "./data/ccs2plus.db", "SQLite database path")
	flag.StringVar(&cfg.ExportDir, "export-dir", "./exports", "directory used for generated daily Excel files")
	flag.StringVar(&cfg.ExportTimezone, "export-timezone", "Asia/Bangkok", "IANA timezone used for daily exports")
	flag.StringVar(&cfg.GoogleDriveCredentials, "gdrive-credentials", "", "path to Google service account JSON for Drive uploads")
	flag.StringVar(&cfg.GoogleDriveFolderID, "gdrive-folder-id", "", "Google Drive folder ID for daily Excel uploads")
	flag.StringVar(&cfg.DashboardDir, "dashboard-dir", "", "directory containing built Flutter dashboard web assets")
	flag.Parse()

	if cfg.SerialBaud <= 0 {
		return AppConfig{}, fmt.Errorf("invalid baud rate: %d", cfg.SerialBaud)
	}

	if cfg.WriteDelayMS < 0 {
		return AppConfig{}, fmt.Errorf("invalid write delay: %d", cfg.WriteDelayMS)
	}

	if cfg.DBPath == "" {
		return AppConfig{}, fmt.Errorf("db path must not be empty")
	}

	if (cfg.GoogleDriveCredentials == "") != (cfg.GoogleDriveFolderID == "") {
		return AppConfig{}, fmt.Errorf("gdrive credentials and folder id must be configured together")
	}

	return cfg, nil
}
