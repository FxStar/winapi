package gdi

import (
	"bytes"
	"image"
	"syscall"
	"unsafe"

	"golang.org/x/image/bmp"

	. "github.com/leibnewton/winapi"
)

var (
	modgdi32 = syscall.NewLazyDLL("gdi32.dll")

	procCreateCompatibleDC = modgdi32.NewProc("CreateCompatibleDC")
	procGetObjectA         = modgdi32.NewProc("GetObjectA")
	procGetObjectW         = modgdi32.NewProc("GetObjectW")
	procSelectObject       = modgdi32.NewProc("SelectObject")
	procDeleteObject       = modgdi32.NewProc("DeleteObject")

	procCreateCompatibleBitmap = modgdi32.NewProc("CreateCompatibleBitmap")
	procCreateDIBSection       = modgdi32.NewProc("CreateDIBSection")
	procBitBlt                 = modgdi32.NewProc("BitBlt")
	procSetDIBits              = modgdi32.NewProc("SetDIBits")
)

type DIBBITMAPINFO struct {
	bfOffBits uint32
	BITMAPINFO
}

func CreateCompatibleDC(hwnd HWND) (hdc HDC) {
	r0, _, _ := syscall.Syscall(procCreateCompatibleDC.Addr(), 1, uintptr(hwnd), 0, 0)
	hdc = HDC(r0)
	return hdc
}

// reference: https://stackoverflow.com/questions/2886831/win32-c-c-load-image-from-memory-buffer/2901465#2901465
func LoadBitmapFromMemory(img image.Image) (HBITMAP, int, int, error) {
	var buf bytes.Buffer
	if err := bmp.Encode(&buf, img); err != nil {
		return 0, 0, 0, err
	}
	size := img.Bounds().Size()
	bmpBuf := buf.Bytes()
	dibBmInfo := (*DIBBITMAPINFO)(unsafe.Pointer(&bmpBuf[10])) // skip first 10B in tagBITMAPFILEHEADER
	pixels := bmpBuf[dibBmInfo.bfOffBits:]
	h, err := CreateDIBSection(0, &dibBmInfo.BITMAPINFO, DIB_RGB_COLORS, 0, 0, 0)
	if err != nil {
		return 0, 0, 0, err
	}
	err = SetDIBits(0, h, 0, dibBmInfo.Header.Height, pixels, &dibBmInfo.BITMAPINFO, DIB_RGB_COLORS)
	return h, size.X, size.Y, err
}

func GetObjectA(hgdiobj HANDLE, cbBuffer uintptr, object uintptr) (size uint32) {
	r0, _, _ := syscall.Syscall(procGetObjectA.Addr(), 3, uintptr(hgdiobj), uintptr(cbBuffer), object)
	size = uint32(r0)
	return size
}

func GetObjectW(hgdiobj HANDLE, cbBuffer uintptr, object uintptr) (size uint32) {
	r0, _, _ := syscall.Syscall(procGetObjectW.Addr(), 3, uintptr(hgdiobj), uintptr(cbBuffer), object)
	size = uint32(r0)
	return size
}

func SelectObject(hdc HDC, hgdiobj HANDLE) HDC {
	r0, _, _ := syscall.Syscall(procSelectObject.Addr(), 2, uintptr(hdc), uintptr(hgdiobj), 0)
	return HDC(r0)
}

func DeleteObject(hgdiobj HANDLE) HANDLE {
	r0, _, _ := syscall.Syscall(procDeleteObject.Addr(), 1, uintptr(hgdiobj), 0, 0)
	return HANDLE(r0)
}

func CreateCompatibleBitmap(hdc HDC, width, height uintptr) (hbitmap HANDLE) {
	r0, _, _ := syscall.Syscall(procCreateCompatibleBitmap.Addr(), 3, uintptr(hdc), uintptr(width), uintptr(height))
	return HANDLE(r0)
}

func CreateDIBSection(hdc HDC, pbmi *BITMAPINFO, iUsage uint, ppvBits uintptr, hSection uint32, dwOffset uint32) (HANDLE, error) {
	r0, _, err := syscall.Syscall6(procCreateDIBSection.Addr(), 6, uintptr(hdc), uintptr(unsafe.Pointer(pbmi)), uintptr(iUsage), ppvBits, uintptr(hSection), uintptr(dwOffset))
	if r0 == 0 {
		return 0, err
	}
	return HANDLE(r0), nil
}

func SetDIBits(hdc HDC, hbm HBITMAP, start, cLines int32, pixels []byte, pbmi *BITMAPINFO, colorUse uint) error {
	r0, _, err := procSetDIBits.Call(uintptr(hdc), uintptr(hbm), uintptr(start), uintptr(cLines), uintptr(unsafe.Pointer(&pixels[0])), uintptr(unsafe.Pointer(pbmi)), uintptr(colorUse))
	if r0 == 0 {
		return err
	}
	return nil
}

func BitBlt(hdc HDC, nXDest, nYDest, nWidth, nHeight int, hdcSrc HDC, nXSrc, nYSrc int, dwRop uint32) bool {
	r0, _, _ := syscall.Syscall9(procBitBlt.Addr(), 9, uintptr(hdc), uintptr(nXDest), uintptr(nYDest), uintptr(nWidth), uintptr(nHeight), uintptr(hdcSrc), uintptr(nXSrc), uintptr(nYSrc), uintptr(dwRop))
	return r0 != 0
}
