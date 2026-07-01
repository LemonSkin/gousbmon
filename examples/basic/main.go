package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/LemonSkin/gousbmon"
)

func main() {
	monitor, err := gousbmon.NewMonitor()

	// New Filter API: Create a Filter and add criteria
	// f := gousbmon.NewFilter().MatchVendorID("18d1")
	// monitor, err := gousbmon.NewMonitor(gousbmon.WithFilters(f))

	// AND filters together: Add multiple criteria to a single filter
	// Match devices with USB interface "HID" AND vendor ID "18D1"
	// f := gousbmon.NewFilter().MatchUSBInterfaces("HID").MatchVendorID("18D1")
	// monitor, err := gousbmon.New(gousbmon.WithFilters(f))

	// OR filters together: Add multiple filters to match different criteria
	// Match devices with USB interface "HID" OR vendor ID "18D1"
	// f1 := gousbmon.NewFilter().MatchUSBInterfaces("HID")
	// f2 := gousbmon.NewFilter().MatchVendorID("18D1")
	// monitor, err := gousbmon.New(gousbmon.WithFilters(f1, f2))

	// AND/OR wombocombo: Combine the above to create more complex filters
	// Match devices with vendor ID 18D1 OR devices with vendor ID 1234 and model ID 5678
	// f1 := gousbmon.NewFilter().MatchVendorID("18D1")
	// f2 := gousbmon.NewFilter().MatchVendorID("1234").MatchModelID("5678")
	// monitor, err := gousbmon.New(gousbmon.WithFilters(f1, f2))
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

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	monitor.StopMonitoring()
	log.Println("Monitoring stopped.")
}
