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
	// You can also add a new filter, e.g.: Filter devices with VID=18D1
	// monitor, err := gousbmon.New(filter.MatchVendorID("18D1"))

	// OR filters together, e.g.: Filter devices with Interface=HID || VID=18D1
	// monitor, err := gousbmon.New(filter.MatchUSBInterfaces("HID"), filter.MatchVendorID("18D1"))

	// AND filters together, e.g.: Filter devices with Interface=HID && VID=18D1
	// monitor, err := gousbmon.New(filter.MatchAll(filter.MatchUSBInterfaces("HID"), filter.MatchVendorID("18D1")))

	// AND/OR wombocombo, e.g.: Filter devices with (VID=18D1) || (VID=1234 && ModelID=5678)
	// monitor, err := gousbmon.New(filter.MatchVendorID("18D1"), filter.MatchAll(filter.MatchVendorID("1234"), filter.MatchModelID("5678")))
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

	// Start background monitor and only detect new connections
	// err = monitor.StartMonitoring(onConnect, nil)

	// Or with a custom polling interval, e.g. 1 second:
	// err = monitor.StartMonitoringWithInterval(onConnect, onDisconnect, 1*time.Second)
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
