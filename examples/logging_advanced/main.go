package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/LemonSkin/gousbmon"
)

func main() {
	// Create an application-level logger for your own code.
	appLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Example: Use the app logger for your application's own logs.
	appLogger.Info("Starting USB monitoring application")

	// Create a separate handler specifically for gousbmon's debug logs.
	gousbmonHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	monitor, err := gousbmon.NewMonitor(gousbmon.WithHandler(gousbmonHandler))

	if err != nil {
		log.Fatalf("failed to create USB monitor: %v", err)
	}

	onConnect := func(deviceID string, deviceInfo gousbmon.DeviceInfo) {
		fmt.Printf("Connected: %s - %s (%s - %s)\n",
			deviceID,
			deviceInfo.IDModel,
			deviceInfo.IDModelID,
			deviceInfo.IDVendorID,
		)
	}

	onDisconnect := func(deviceID string, deviceInfo gousbmon.DeviceInfo) {
		fmt.Printf("Disconnected: %s - %s (%s - %s)\n",
			deviceID,
			deviceInfo.IDModel,
			deviceInfo.IDModelID,
			deviceInfo.IDVendorID,
		)
	}

	err = monitor.StartMonitoring(onConnect, onDisconnect)

	if err != nil {
		log.Fatalf("failed to start monitoring: %v", err)
	}

	log.Println("Monitoring USB devices... Press Ctrl+C to stop.")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	monitor.StopMonitoring()
	log.Println("Monitoring stopped.")
}
