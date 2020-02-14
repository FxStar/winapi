package main

import (
	"log"

	"github.com/leibnewton/winapi"
	"github.com/leibnewton/winapi/gdi"
	"github.com/leibnewton/winapi/winspool"
)

func main() {
	// TODO: EnumPrinters
	prn, err := winspool.CreateDC("Microsoft XPS Document Writer", nil)
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
