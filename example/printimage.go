package main

import (
	"fmt"
	"image/png"
	"log"
	"os"

	"github.com/leibnewton/winapi"
	"github.com/leibnewton/winapi/gdi"
	"github.com/leibnewton/winapi/winspool"
)

// 打印机设置要点：
// 1.纸张大小：40mm*30mm（2mm间隔）
// 2.四个方向边距设置为0.
// 3.分辨率为300dpi时，使用472*354分辨率的图片
//   分辨率为200dpi时，使用315*236分辨率的图片
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

	//Method 1: use LoadImageFromFile to load disk image
	//img, err := winapi.LoadImageFromFile(`sample.bmp`, winapi.IMAGE_BITMAP)
	//imgWidth := 480 // width and height should be the real values of the picture
	//imgHeight := 360

	// Method 2:
	f, err := os.Open("sample300.png")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	pngimg, err := png.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	img, imgWidth, imgHeight, err := gdi.LoadBitmapFromMemory(pngimg)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("create bmp image success, width:%d, height:%d", imgWidth, imgHeight)
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

	// Alternative: use StretchBlt to scale
	gdi.BitBlt(prn, 0, 0, imgWidth, imgHeight,
		hdcMem, 0, 0, gdi.SRCCOPY)
}
