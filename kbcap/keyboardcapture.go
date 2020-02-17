package kbcap

// based on: https://gist.github.com/obonyojimmy/52d836a1b31e2fc914d19a81bd2e0a1b
//           https://gist.github.com/sbarratt/3077d5f51288b39665350dc2b9e19694

import (
	"bytes"
	"log"
	"syscall"
	"time"
	"unsafe"

	"github.com/leibnewton/winapi"
)

// String returns a human-friendly display name of the hotkey
// such as "Hotkey[Id: 1, Alt+Ctrl+O]"
var (
	user32 = syscall.NewLazyDLL("user32.dll")

	procSetWindowsHookEx    = user32.NewProc("SetWindowsHookExW")
	procCallNextHookEx      = user32.NewProc("CallNextHookEx")
	procUnhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")
)

const (
	WH_KEYBOARD_LL = 13
	WH_KEYBOARD    = 2
	WM_KEYDOWN     = 256
	WM_SYSKEYDOWN  = 260
	WM_KEYUP       = 257
	WM_SYSKEYUP    = 261
	WM_KEYFIRST    = 256
	WM_KEYLAST     = 264
	PM_NOREMOVE    = 0x000
	PM_REMOVE      = 0x001
	PM_NOYIELD     = 0x002
	WM_LBUTTONDOWN = 513
	WM_RBUTTONDOWN = 516
	NULL           = 0
)

type (
	DWORD     = uint32
	WPARAM    uintptr
	LPARAM    uintptr
	LRESULT   uintptr
	HANDLE    uintptr
	HINSTANCE HANDLE
	HHOOK     HANDLE
	HWND      HANDLE
)

type HOOKPROC func(int, WPARAM, LPARAM) LRESULT

type KBDLLHOOKSTRUCT struct {
	VkCode      DWORD
	ScanCode    DWORD
	Flags       DWORD
	Time        DWORD
	DwExtraInfo uintptr
}

func SetWindowsHookEx(idHook int, lpfn HOOKPROC, hMod HINSTANCE, dwThreadId DWORD) (HHOOK, error) {
	ret, _, err := procSetWindowsHookEx.Call(uintptr(idHook), uintptr(syscall.NewCallback(lpfn)), uintptr(hMod), uintptr(dwThreadId))
	if ret == 0 {
		return 0, err
	}
	return HHOOK(ret), nil
}

func (hhk HHOOK) CallNextHookEx(nCode int, wParam WPARAM, lParam LPARAM) LRESULT {
	ret, _, _ := procCallNextHookEx.Call(uintptr(hhk), uintptr(nCode), uintptr(wParam), uintptr(lParam))
	return LRESULT(ret)
}

func (hhk *HHOOK) UnhookWindowsHookEx() bool {
	ret, _, _ := procUnhookWindowsHookEx.Call(uintptr(*hhk))
	*hhk = 0
	return ret != 0
}

func GetAnyMessage() { // block
	winapi.GetMessage(nil, 0, 0, 0)
}

const (
	kShiftKey   = 16
	kCapitalKey = 20
)

func CodeToChar(hookStruct *KBDLLHOOKSTRUCT) (byte, bool) {
	keyStates, _ := winapi.GetKeyboardState()
	keyStates[kShiftKey] = byte(winapi.GetKeyState(winapi.VK_SHIFT))
	keyStates[kCapitalKey] = byte(winapi.GetKeyState(winapi.VK_CAPITAL))
	var char uint16
	n := winapi.ToAscii(hookStruct.VkCode, hookStruct.ScanCode, keyStates, &char, 0)
	return byte(char), n == 1
}

var MaxUpdateInterval time.Duration = 3 * time.Second
var Debug = false

func MonitorKeyboard(callback func(string)) error {
	var buf bytes.Buffer
	buf.Grow(128)
	lastUpdate := time.Now()

	var keyboardHook HHOOK
	keyboardHook, err := SetWindowsHookEx(WH_KEYBOARD_LL,
		(HOOKPROC)(func(nCode int, wparam WPARAM, lparam LPARAM) LRESULT {
			if nCode == 0 && wparam == WM_KEYDOWN {
				kbdstruct := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lparam))
				if b, ok := CodeToChar(kbdstruct); ok {
					if time.Since(lastUpdate) > MaxUpdateInterval && buf.Len() > 0 {
						buf.Reset()
						if Debug {
							log.Printf("exceed max update interval of %v, reset", MaxUpdateInterval)
						}
					}
					lastUpdate = time.Now()
					if b != '\n' && b != '\r' {
						buf.WriteByte(b)
						if Debug {
							log.Printf("key pressed: '%c' l=%d", b, buf.Len())
						}
					} else if buf.Len() > 0 {
						if callback != nil {
							callback(buf.String())
						}
						if Debug {
							log.Printf("get line: '%s'", buf.String())
						}
						buf.Reset()
					}
				}
			}
			return keyboardHook.CallNextHookEx(nCode, wparam, lparam)
		}), 0, 0)
	if err != nil {
		return err
	}

	log.Printf("keyboard monitoring...")
	go func() {
		defer keyboardHook.UnhookWindowsHookEx()
		for {
			GetAnyMessage() // may block infinitely, while PeekMessage will consume more CPU or less responsive.
		}
	}()
	return nil
}
