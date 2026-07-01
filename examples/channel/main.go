package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/LemonSkin/gousbmon"
)

func main() {
	monitor, err := gousbmon.NewMonitor()
	if err != nil {
		log.Fatalf("failed to create USB monitor: %v", err)
	}

	connectedChan := make(chan gousbmon.DeviceInfo)
	disconnectedChan := make(chan gousbmon.DeviceInfo)

	onConnect := func(deviceID string, deviceInfo gousbmon.DeviceInfo) {
		connectedChan <- deviceInfo
	}
	onDisconnect := func(deviceID string, deviceInfo gousbmon.DeviceInfo) {
		disconnectedChan <- deviceInfo
	}

	err = monitor.StartMonitoring(onConnect, onDisconnect)
	if err != nil {
		log.Fatalf("failed to start monitoring: %v", err)
	}

	log.Println("Monitoring USB devices... Press Ctrl+C to stop.")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case info := <-connectedChan:
			handleConnected(info)
		case info := <-disconnectedChan:
			handleDisconnected(info)
		case <-sigCh:
			monitor.StopMonitoring()
			log.Println("Monitoring stopped.")
			return
		}
	}
}

func handleConnected(info gousbmon.DeviceInfo) {
	log.Printf("Connected: %s (%s - %s)\n",
		info.IDModel,
		info.IDModelID,
		info.IDVendorID,
	)
}

func handleDisconnected(info gousbmon.DeviceInfo) {
	log.Printf("Disconnected: %s (%s - %s)\n",
		info.IDModel,
		info.IDModelID,
		info.IDVendorID,
	)
}
