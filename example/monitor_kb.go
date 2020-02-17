package main

import (
	"log"

	"github.com/leibnewton/winapi/kbcap"
)

func main() {
	err := kbcap.MonitorKeyboard(nil)
	if err != nil {
		log.Fatal(err)
	}
	select {}
}
