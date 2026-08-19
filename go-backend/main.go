package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"ccs2plus/go-backend/internal/config"
	"ccs2plus/go-backend/internal/httpapi"
	"ccs2plus/go-backend/internal/reporting"
	"ccs2plus/go-backend/internal/roommap"
	"ccs2plus/go-backend/internal/store"
	"ccs2plus/go-backend/internal/transport"
)

func main() {
	cfg, err := config.LoadAppConfig()
	if err != nil {
		log.Fatalf("load app config: %v", err)
	}

	mappings := []roommap.RoomMapping{
		{
			RoomID:       "1205",
			ControllerID: "ctrl-1205",
			Address:      0x03,
			CardIndex:    3,
			PortIndex:    5,
		},
	}

	if cfg.RoomMapPath != "" {
		roomCfg, err := config.LoadRoomMapFile(cfg.RoomMapPath)
		if err != nil {
			log.Fatalf("load room map: %v", err)
		}
		mappings = roomCfg.Rooms
	}

	repo := roommap.NewMemoryRepository(mappings)
	dbStore, err := store.Open(cfg.DBPath, repo)
	if err != nil {
		log.Fatalf("open sqlite store: %v", err)
	}
	defer func() {
		if err := dbStore.Close(); err != nil {
			log.Printf("close sqlite store: %v", err)
		}
	}()

	location, err := time.LoadLocation(cfg.ExportTimezone)
	if err != nil {
		log.Fatalf("load export timezone: %v", err)
	}

	if cfg.GoogleDriveCredentials != "" {
		if _, err := dbStore.SaveGoogleDriveConfig(store.GoogleDriveConfig{
			CredentialsPath: cfg.GoogleDriveCredentials,
			FolderID:        cfg.GoogleDriveFolderID,
			Enabled:         true,
		}); err != nil {
			log.Fatalf("save google drive config: %v", err)
		}
	}

	exporter := reporting.NewDailyExporter(dbStore, cfg.ExportDir, location, nil)
	exporter.Start(context.Background())

	var writer transport.Writer = &transport.MemoryWriter{}

	if cfg.SerialDevice != "" {
		serialWriter, err := transport.NewSerialWriter(cfg.SerialDevice, cfg.SerialBaud, time.Duration(cfg.WriteDelayMS)*time.Millisecond)
		if err != nil {
			log.Fatalf("open serial writer: %v", err)
		}
		defer func() {
			if err := serialWriter.Close(); err != nil {
				log.Printf("close serial writer: %v", err)
			}
		}()
		svc := transport.NewService(repo, serialWriter)
		server := httpapi.NewServerWithStore(svc, repo, dbStore)
		server.SetDashboardDir(cfg.DashboardDir)
		log.Printf("master API listening on %s, serial=%s baud=%d", cfg.ListenAddr, cfg.SerialDevice, cfg.SerialBaud)
		log.Fatal(http.ListenAndServe(cfg.ListenAddr, server.Handler()))
	}

	svc := transport.NewService(repo, writer)
	server := httpapi.NewServerWithStore(svc, repo, dbStore)
	server.SetDashboardDir(cfg.DashboardDir)
	log.Printf("master API listening on %s in memory-writer mode", cfg.ListenAddr)
	log.Fatal(http.ListenAndServe(cfg.ListenAddr, server.Handler()))
}
