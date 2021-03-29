package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/leibnewton/winapi/winspool"
)

var rawFile = flag.String("file", "document.raw", "file path")
var pageCount = flag.Int("count", 1, "page count")

func Fatal(v ...interface{}) {
	log.Print(v...)
	fmt.Scanln()
	os.Exit(1)
}

func main() {
	flag.Parse()
	defer fmt.Scanln()
	defer fmt.Print("\nDone. Print any key to quit...")
	// EnumPrinters
	printers, err := winspool.EnumPrinters4()
	if err != nil {
		Fatal(err)
	}
	if len(printers) == 0 {
		log.Printf("NO PRINTERS FOUND.")
		return
	}
	fmt.Printf("Found printers:\n")
	for i, p := range printers {
		fmt.Printf("  [%d] %s\t(local=%v, online=%v)\n", i, p.GetPrinterName(), p.IsLocal(), p.IsOnline())
	}
	fmt.Printf("Which printer to select [0-%d]: ", len(printers)-1)
	index := 0
	if _, err = fmt.Scanln(&index); err != nil {
		Fatal(err)
	}
	printerName := printers[index].GetPrinterName()

	f, err := os.Open(*rawFile)
	if err != nil {
		Fatal("open failed", err)
	}
	defer f.Close()

	hPrinter, err := winspool.OpenPrinter(printerName)
	if err != nil {
		Fatal("OpenPrinter", err)
	}
	defer hPrinter.ClosePrinter() // 关闭打印机时, 没打印出的页会丢弃.

	dwJob, err := hPrinter.StartDoc("Printing Raw File...")
	if err != nil {
		Fatal("StartDoc", err)
	}
	log.Printf("get job: %d", dwJob)
	defer hPrinter.EndDoc()

	err = hPrinter.SetJobCommand(dwJob, winspool.JOB_CONTROL_RETAIN)
	if err != nil {
		hPrinter.SetJobCommand(dwJob, winspool.JOB_CONTROL_RELEASE)
		Fatal("retain job failed", err)
	}
	log.Printf("job: %d retained", dwJob)

	for p := 0; p < *pageCount; p++ {
		err = hPrinter.StartPage()
		if err != nil {
			Fatal("StartPage", err)
		}

		written, err := io.Copy(hPrinter, f)
		if err != nil {
			log.Printf("#%d write failed, %d written, reason: %v", p, written, err)
		} else {
			log.Printf("#%d write done", p)
		}
		err = hPrinter.EndPage()
		if err != nil {
			Fatal("EndPage", err)
		}
	}
}
