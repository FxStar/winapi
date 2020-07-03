package main

import (
	"flag"
	"fmt"
	"image/png"
	"io"
	"log"
	"os"
	"time"

	"github.com/leibnewton/winapi"
	"github.com/leibnewton/winapi/gdi"
	"github.com/leibnewton/winapi/winspool"
)

var shouldCheck = flag.Bool("check", false, "check job status")
var pageCount = flag.Int("count", 1, "page count")

func Fatal(v ...interface{}) {
	log.Print(v...)
	fmt.Scanln()
	os.Exit(1)
}

// 打印机设置要点：
// 1.纸张大小：40mm*30mm（2mm间隔）
// 2.四个方向边距设置为0.
// 3.分辨率为300dpi时，使用472*354分辨率的图片
//   分辨率为200dpi时，使用315*236分辨率的图片
func main() {
	flag.Parse()
	defer fmt.Scanln()
	if *shouldCheck {
		logf, err := os.OpenFile("spooljob.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Printf("open log file failed: %v", err)
			return
		}
		defer logf.Close()
		log.SetOutput(io.MultiWriter(logf, os.Stdout))
	}
	log.Printf("============== begin ==============>")

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

	// Printing
	prn, err := winspool.CreateDC(printerName, nil)
	if err != nil {
		Fatal(err)
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
		Fatal(err)
	}
	defer f.Close()
	pngimg, err := png.Decode(f)
	if err != nil {
		Fatal(err)
	}
	img, imgWidth, imgHeight, err := gdi.LoadBitmapFromMemory(pngimg)
	if err != nil {
		Fatal(err)
	}
	log.Printf("create bmp image success, width:%d, height:%d", imgWidth, imgHeight)
	defer gdi.DeleteObject(img) //When you are finished using a bitmap you loaded without specifying the LR_SHARED flag, you can release its associated memory by calling DeleteObject.
	gdi.SelectObject(hdcMem, img)

	hPrinter, err := winspool.OpenPrinter(printerName)
	if err != nil {
		Fatal(err)
	}
	defer hPrinter.ClosePrinter()
	log.Printf("check non-existent job 0")
	checkJob(hPrinter, 0)

	jobId, err := prn.StartDoc("Printing Picture...")
	if err != nil {
		Fatal(err)
	}
	for i:=0; i<*pageCount; i++ {
		if err = prn.StartPage(); err != nil {
			Fatal(err)
		}
		// Alternative: use StretchBlt to scale
		gdi.BitBlt(prn, 0, 0, imgWidth, imgHeight,
			hdcMem, 0, 0, gdi.SRCCOPY)
		prn.EndPage()

		log.Printf("issued page #%d", i)
		checkJob(hPrinter, jobId)
	}
	prn.EndDoc()

	log.Printf("====== end doc ======")
	if *shouldCheck {
		if err := hPrinter.SetJobCommand(jobId, winspool.JOB_CONTROL_RETAIN); err != nil {
			Fatal("retain job failed:", err)
		}
		log.Printf(">>> job retained")
		start := time.Now()
		for time.Since(start) < time.Hour {
			checkJob(hPrinter, jobId)
			time.Sleep(2 * time.Second)
		}
	}
}

// https://github.com/google/cloud-print-connector/blob/master/winspool/winspool.go
// L913～L948
func checkJob(hPrinter winspool.HANDLE, jobId int32) error {
	if !*shouldCheck{
		return nil
	}

	ji, err := hPrinter.GetJob(jobId)
	if err != nil {
		log.Printf("get job failed: %v", err)
		return err
	}
	statusList := []struct{
		code uint32
		name string
	}{
		{winspool.JOB_STATUS_PAUSED, "PAUSED"},
		{winspool.JOB_STATUS_ERROR, "ERROR"},
		{winspool.JOB_STATUS_DELETING, "DELETING"},
		{winspool.JOB_STATUS_SPOOLING, "SPOOLING"},
		{winspool.JOB_STATUS_PRINTING, "PRINTING"},
		{winspool.JOB_STATUS_OFFLINE, "OFFLINE"},
		{winspool.JOB_STATUS_PAPEROUT, "PAPEROUT"},
		{winspool.JOB_STATUS_PRINTED, "PRINTED"},
		{winspool.JOB_STATUS_DELETED, "DELETED"},
		{winspool.JOB_STATUS_BLOCKED_DEVQ, "BLOCKED_DEVQ"},
		{winspool.JOB_STATUS_USER_INTERVENTION, "USER_INTERVENTION"},
		{winspool.JOB_STATUS_RESTART, "RESTART"},
		{winspool.JOB_STATUS_COMPLETE, "COMPLETE"},
		{winspool.JOB_STATUS_RETAINED, "RETAINED"},
		{winspool.JOB_STATUS_RENDERING_LOCALLY, "RENDERING_LOCALLY"},
	}
	var curstatus []string
	detectError := false
	complete := false
	for _, item := range statusList {
		if (ji.GetStatus() & item.code) == item.code {
			curstatus = append(curstatus, item.name)
			complete = complete || (item.code == winspool.JOB_STATUS_PRINTED) || (item.code == winspool.JOB_STATUS_COMPLETE)
			detectError = detectError || (item.code == winspool.JOB_STATUS_ERROR)
		}
	}
	log.Printf("  printed: %d/%d, status: %v", ji.GetPagesPrinted(),ji.GetTotalPages(), curstatus)
	if complete {
		if err := hPrinter.SetJobCommand(jobId, winspool.JOB_CONTROL_RELEASE); err != nil {
			Fatal("release job failed:", err)
		}
		Fatal("detect complete")
	}
	if detectError {
		if err := hPrinter.SetJobCommand(jobId, winspool.JOB_CONTROL_RESUME); err != nil {
			log.Printf("resume job failed: %v", err)
		}else{
			log.Printf("detect error, resume job")
		}
	}
	return nil
}
