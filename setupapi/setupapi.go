// adapt setupapi Functions
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
// OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
// DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
// TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
// OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package setupapi

// reference: https://github.com/distatus/battery battery_windows.go

import (
	"syscall"
	"unsafe"

	"github.com/leibnewton/winapi/winspool"
)

var (
	setupapi                         = syscall.NewLazyDLL("setupapi.dll")
	setupDiGetClassDevsW             = setupapi.NewProc("SetupDiGetClassDevsW")
	setupDiEnumDeviceInterfaces      = setupapi.NewProc("SetupDiEnumDeviceInterfaces")
	setupDiGetDeviceInterfaceDetailW = setupapi.NewProc("SetupDiGetDeviceInterfaceDetailW")
	setupDiDestroyDeviceInfoList     = setupapi.NewProc("SetupDiDestroyDeviceInfoList")
)

// Flags controlling what is included in the device information set built by SetupDiGetClassDevs
const (
	DIGCF_DEFAULT         = 0x00000001 // only valid with DIGCF_DEVICEINTERFACE
	DIGCF_PRESENT         = 0x00000002
	DIGCF_ALLCLASSES      = 0x00000004
	DIGCF_PROFILE         = 0x00000008
	DIGCF_DEVICEINTERFACE = 0x00000010
)

type guid struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

type spDeviceInterfaceData struct {
	cbSize             uint32
	InterfaceClassGuid guid
	Flags              uint32
	Reserved           uint
}

type spDevInfoData struct {
	cbSize    uint32
	ClassGuid guid
	DevInst   uint32
	Reserved  uint
}

var guidDevicePrinter = guid{
	0x28d78fad,
	0x5a12,
	0x11D1,
	[8]byte{0xae, 0x5b, 0x00, 0x00, 0xf8, 0x03, 0xa8, 0xc2},
}

type HDEVINFO struct {
	h          uintptr
	devicePath string
	devInfo    spDevInfoData
}

func (hDevs *HDEVINFO) DevicePath() string {
	return hDevs.devicePath
}

func GetClassDevs() (*HDEVINFO, error) {
	r1, _, err := setupDiGetClassDevsW.Call(uintptr(unsafe.Pointer(&guidDevicePrinter)), 0, 0, DIGCF_PRESENT|DIGCF_DEVICEINTERFACE)
	if r1 == ^uintptr(0) { // -1
		return nil, err
	}
	return &HDEVINFO{h: r1}, nil
}

func (hDevs *HDEVINFO) DestroyDeviceInfoList() error {
	r1, _, err := setupDiDestroyDeviceInfoList.Call(uintptr(hDevs.h))
	if r1 == 0 { // BOOL
		return err
	}
	hDevs.h = 0
	return nil
}

func (hDevs *HDEVINFO) EnumDeviceInterfaces(idx int) (bool, error) {
	var did spDeviceInterfaceData
	did.cbSize = uint32(unsafe.Sizeof(did))
	r1, _, err := setupDiEnumDeviceInterfaces.Call(uintptr(hDevs.h), 0, uintptr(unsafe.Pointer(&guidDevicePrinter)), uintptr(idx), uintptr(unsafe.Pointer(&did)))
	if r1 == 0 {
		if err == winspool.ERROR_NO_MORE_ITEMS {
			err = nil
		}
		return false, err
	}
	var cbRequired uint32
	_, _, err = setupDiGetDeviceInterfaceDetailW.Call(uintptr(hDevs.h), uintptr(unsafe.Pointer(&did)), 0, 0, uintptr(unsafe.Pointer(&cbRequired)), 0)
	if err != winspool.ERROR_INSUFFICIENT_BUFFER {
		return false, err
	}

	pDevDetail := make([]uint16, cbRequired/2)
	cbSize := (*uint32)(unsafe.Pointer(&pDevDetail[0]))
	*cbSize = 6
	if unsafe.Sizeof(uint(0)) == 8 {
		*cbSize = 8
	}
	hDevs.devInfo.cbSize = uint32(unsafe.Sizeof(hDevs.devInfo))
	r1, _, err = setupDiGetDeviceInterfaceDetailW.Call(uintptr(hDevs.h), uintptr(unsafe.Pointer(&did)), uintptr(unsafe.Pointer(&pDevDetail[0])), uintptr(cbRequired), uintptr(unsafe.Pointer(&cbRequired)), uintptr(unsafe.Pointer(&hDevs.devInfo)))
	if r1 == 0 {
		return false, err
	}
	hDevs.devicePath = syscall.UTF16ToString(pDevDetail[2:])
	return true, nil
}
