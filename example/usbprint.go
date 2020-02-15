package main

import (
	"log"

	"github.com/leibnewton/winapi/setupapi"
)

func main() {
	hDev, err := setupapi.GetClassDevs()
	if err != nil {
		log.Fatal(err)
	}
	defer hDev.DestroyDeviceInfoList()

	for idx := 0; ; idx++ {
		ok, err := hDev.EnumDeviceInterfaces(idx)
		if err != nil {
			log.Fatal(err)
		}
		if !ok {
			break
		}
		vid, pid, err := hDev.GetVidPid()
		if err != nil {
			log.Printf("get hardwareId failed: %v", err)
		}
		location, err := hDev.GetLocation()
		if err != nil {
			log.Printf("get location failed: %v", err)
		}
		hardwareId, err := hDev.GetHardwareId()
		if err != nil {
			log.Printf("get hardwareId failed: %v", err)
		}
		log.Printf("  [%d] %s VID:%04X PID:%04X location: %s hardwareid: %s",
			idx, hDev.DevicePath(), vid, pid, location, hardwareId)
	}
}
