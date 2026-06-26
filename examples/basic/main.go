package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/LemonSkin/gousbmon"
	"github.com/LemonSkin/gousbmon/device"
)

func main() {
	monitor, err := gousbmon.New()
	// You can also add a new filter, e.g:
	// monitor, err := gousbmon.New(gousbmon.MatchVendorID("18D1"))

	// Or multiple filters, e.g.:
	// monitor, err := gousbmon.New(gousbmon.MatchUSBInterfaces("HID"), gousbmon.MatchVendorID("046d"))
	if err != nil {
		log.Fatalf("failed to create USB monitor: %v", err)
	}

	// Example 1: View currently connected devices
	devices, err := monitor.GetAvailableDevices()
	if err != nil {
		log.Fatalf("failed to get devices: %v", err)
	}

	for id, info := range devices {
		fmt.Printf("%s -- %s (%s - %s)\n", id, info.IDModel, info.IDModelID, info.IDVendorID)
	}

	// Example 2: Monitor for device connections/disconnections
	// Define callback functions
	onConnect := func(deviceID string, deviceInfo device.Info) {
		fmt.Printf("Connected: %s - %s (%s - %s)\n",
			deviceID,
			deviceInfo.IDModel,
			deviceInfo.IDModelID,
			deviceInfo.IDVendorID,
		)
	}

	onDisconnect := func(deviceID string, deviceInfo device.Info) {
		fmt.Printf("Disconnected: %s - %s (%s - %s)\n",
			deviceID,
			deviceInfo.IDModel,
			deviceInfo.IDModelID,
			deviceInfo.IDVendorID,
		)
	}

	// Start the background monitor with a default polling interval of 500ms
	err = monitor.StartMonitoring(onConnect, onDisconnect)

	// Or with a custom polling interval, e.g. 1 second:
	// err = monitor.StartMonitoring(onConnect, onDisconnect, gousbmon.WithInterval(1*time.Second))
	if err != nil {
		log.Fatalf("failed to start monitoring: %v", err)
	}

	log.Println("Monitoring USB devices... Press Ctrl+C to stop.")

	// Wait for interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// Clean shutdown
	monitor.StopMonitoring()
	log.Println("Monitoring stopped.")
}
