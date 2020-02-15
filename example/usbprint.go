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
		log.Printf("  [%d] %s", idx, hDev.DevicePath())
	}
}
