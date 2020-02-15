package main

import (
	"fmt"
	"log"

	"github.com/leibnewton/winapi"
	"github.com/leibnewton/winapi/gdi"
	"github.com/leibnewton/winapi/winspool"
)

func main() {
	// EnumPrinters
	printers, err := winspool.EnumPrinters4()
	if err != nil {
		log.Fatal(err)
	}
	if len(printers) == 0 {
		log.Printf("NO PRINTERS FOUND.")
		return
	}
	log.Printf("Found printers:")
	for i, p := range printers {
		log.Printf("  [%d] %s\t(local=%v, online=%v)", i, p.GetPrinterName(), p.IsLocal(), p.IsOnline())
	}
	fmt.Printf("Which printer to select [0-%d]: ", len(printers)-1)
	index := 0
	if _, err = fmt.Scanln(&index); err != nil {
		log.Fatal(err)
	}
	printerName := printers[index].GetPrinterName()

	// Printing
	prn, err := winspool.CreateDC(printerName, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer prn.DeleteDC()

	hdcMem := gdi.CreateCompatibleDC(winapi.HWND(prn))
	defer hdcMem.DeleteDC()

	img, err := winapi.LoadImageFromFile(`C:\\workspace\\Sandbox\\WinSpoolSample\\bmp\\landscape.bmp`, winapi.IMAGE_BITMAP)
	if err != nil {
		log.Fatal(err)
	}
	defer gdi.DeleteObject(img) //When you are finished using a bitmap you loaded without specifying the LR_SHARED flag, you can release its associated memory by calling DeleteObject.
	gdi.SelectObject(hdcMem, img)

	if _, err = prn.StartDoc("Printing Picture..."); err != nil {
		log.Fatal(err)
	}
	defer prn.EndDoc()
	if err = prn.StartPage(); err != nil {
		log.Fatal(err)
	}
	defer prn.EndPage()

	// use StretchBlt to scale
	imgWidth := 621
	imgHeight := 484
	gdi.BitBlt(prn, 0, 0, imgWidth, imgHeight,
		hdcMem, 0, 0, gdi.SRCCOPY)
}
