package main

import (
	"log"

	"github.com/FxStar/winapi/kbcap"
)

func main() {
	go func() {
		kbcap.Debug = true
		err := kbcap.MonitorKeyboard(func(line string) {
			log.Printf("callback get line: [%s]", line)
		}, func(b byte) {

		})
		if err != nil {
			log.Fatal(err)
		}
	}()
	select {}
}
