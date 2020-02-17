package main

import (
	"log"

	"github.com/leibnewton/winapi/kbcap"
)

func main() {
	err := kbcap.MonitorKeyboard(func(line string) {
		log.Printf("callback get line: [%s]", line)
	})
	if err != nil {
		log.Fatal(err)
	}
	select {}
}
