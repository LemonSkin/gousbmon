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
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	monitor, err := gousbmon.NewMonitor(gousbmon.WithLogger(logger))

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
