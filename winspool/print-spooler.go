package winspool

import (
	"syscall"
	"unsafe"
)

/*
1. call OpenPrinter
2. To begin a print job, call StartDocPrinter.
3. To begin each page, call StartPagePrinter.
4. To write data to a page, call WritePrinter.
5. To end each page, call EndPagePrinter.
6. Repeat 3~5 for as many pages as necessary.
7. To end the print job, call EndDocPrinter.
8. call ClosePrinter
*/

var (
	// openPrinterProc, closePrinterProc in win32.go
	startDocPrinterProc  = winspool.NewProc("StartDocPrinterW")
	startPagePrinterProc = winspool.NewProc("StartPagePrinter")
	writePrinterProc     = winspool.NewProc("WritePrinter")
	endPagePrinterProc   = winspool.NewProc("EndPagePrinter")
	endDocPrinterProc    = winspool.NewProc("EndDocPrinter")
	procGetDefaultPrinterW = winspool.NewProc("GetDefaultPrinterW")
)

// DOCINFO struct.
type DocInfo1 struct {
	pDocName    *uint16
	pOutputFile *uint16
	pDatatype   *uint16
}

func (hPrinter HANDLE) StartDoc(docName string) (int32, error) {
	var docInfo DocInfo1
	var err error
	docInfo.pDocName, err = syscall.UTF16PtrFromString(docName)
	if err != nil {
		return 0, err
	}
	docInfo.pDatatype, err = syscall.UTF16PtrFromString("RAW")
	if err != nil {
		return 0, err
	}

	r1, _, err := startDocPrinterProc.Call(uintptr(hPrinter), 1, uintptr(unsafe.Pointer(&docInfo)))
	if r1 == 0 {
		return 0, err
	}
	return int32(r1), nil
}

func (hPrinter HANDLE) StartPage() error {
	r1, _, err := startPagePrinterProc.Call(uintptr(hPrinter))
	if r1 == 0 {
		return err
	}
	return nil
}

func (hPrinter HANDLE) EndPage() error {
	r1, _, err := endPagePrinterProc.Call(uintptr(hPrinter))
	if r1 == 0 {
		return err
	}
	return nil
}

func (hPrinter HANDLE) EndDoc() error {
	r1, _, err := endDocPrinterProc.Call(uintptr(hPrinter))
	if r1 == 0 {
		return err
	}
	return nil
}

func (hPrinter HANDLE) Write(data []byte) (int, error) {
	var written uint32
	r1, _, err := writePrinterProc.Call(uintptr(hPrinter),
		uintptr(unsafe.Pointer(&data[0])), uintptr(len(data)), uintptr(unsafe.Pointer(&written)))
	if r1 == 0 {
		return 0, err
	}
	return int(written), nil
}

func GetDefaultPrinter(buf *uint16, bufN *uint32) (err error) {
	r1, _, e1 := syscall.Syscall(procGetDefaultPrinterW.Addr(), 2, uintptr(unsafe.Pointer(buf)), uintptr(unsafe.Pointer(bufN)), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func Default() (string, error) {
	b := make([]uint16, 3)
	n := uint32(len(b))
	err := GetDefaultPrinter(&b[0], &n)
	if err != nil {
		if err != syscall.ERROR_INSUFFICIENT_BUFFER {
			return "", err
		}
		b = make([]uint16, n)
		err = GetDefaultPrinter(&b[0], &n)
		if err != nil {
			return "", err
		}
	}
	return syscall.UTF16ToString(b), nil
}