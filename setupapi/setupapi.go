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
	"errors"
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"github.com/FxStar/winapi"
	"github.com/FxStar/winapi/winspool"
)

var (
	setupapi                          = syscall.NewLazyDLL("setupapi.dll")
	setupDiGetClassDevsW              = setupapi.NewProc("SetupDiGetClassDevsW")
	setupDiEnumDeviceInterfaces       = setupapi.NewProc("SetupDiEnumDeviceInterfaces")
	setupDiGetDeviceInterfaceDetailW  = setupapi.NewProc("SetupDiGetDeviceInterfaceDetailW")
	setupDiDestroyDeviceInfoList      = setupapi.NewProc("SetupDiDestroyDeviceInfoList")
	setupDiGetDeviceRegistryPropertyW = setupapi.NewProc("SetupDiGetDeviceRegistryPropertyW")
)

// Flags controlling what is included in the device information set built by SetupDiGetClassDevs
const (
	DIGCF_DEFAULT         = 0x00000001 // only valid with DIGCF_DEVICEINTERFACE
	DIGCF_PRESENT         = 0x00000002
	DIGCF_ALLCLASSES      = 0x00000004
	DIGCF_PROFILE         = 0x00000008
	DIGCF_DEVICEINTERFACE = 0x00000010
)

// Device registry property codes
const (
	SPDRP_DEVICEDESC                  = 0x00000000 // DeviceDesc (R/W)
	SPDRP_HARDWAREID                  = 0x00000001 // HardwareID (R/W)
	SPDRP_COMPATIBLEIDS               = 0x00000002 // CompatibleIDs (R/W)
	SPDRP_UNUSED0                     = 0x00000003 // unused
	SPDRP_SERVICE                     = 0x00000004 // Service (R/W)
	SPDRP_UNUSED1                     = 0x00000005 // unused
	SPDRP_UNUSED2                     = 0x00000006 // unused
	SPDRP_CLASS                       = 0x00000007 // Class (R--tied to ClassGUID)
	SPDRP_CLASSGUID                   = 0x00000008 // ClassGUID (R/W)
	SPDRP_DRIVER                      = 0x00000009 // Driver (R/W)
	SPDRP_CONFIGFLAGS                 = 0x0000000A // ConfigFlags (R/W)
	SPDRP_MFG                         = 0x0000000B // Mfg (R/W)
	SPDRP_FRIENDLYNAME                = 0x0000000C // FriendlyName (R/W)
	SPDRP_LOCATION_INFORMATION        = 0x0000000D // LocationInformation (R/W)
	SPDRP_PHYSICAL_DEVICE_OBJECT_NAME = 0x0000000E // PhysicalDeviceObjectName (R)
	SPDRP_CAPABILITIES                = 0x0000000F // Capabilities (R)
	SPDRP_UI_NUMBER                   = 0x00000010 // UiNumber (R)
	SPDRP_UPPERFILTERS                = 0x00000011 // UpperFilters (R/W)
	SPDRP_LOWERFILTERS                = 0x00000012 // LowerFilters (R/W)
	SPDRP_BUSTYPEGUID                 = 0x00000013 // BusTypeGUID (R)
	SPDRP_LEGACYBUSTYPE               = 0x00000014 // LegacyBusType (R)
	SPDRP_BUSNUMBER                   = 0x00000015 // BusNumber (R)
	SPDRP_ENUMERATOR_NAME             = 0x00000016 // Enumerator Name (R)
	SPDRP_SECURITY                    = 0x00000017 // Security (R/W, binary form)
	SPDRP_SECURITY_SDS                = 0x00000018 // Security (W, SDS form)
	SPDRP_DEVTYPE                     = 0x00000019 // Device Type (R/W)
	SPDRP_EXCLUSIVE                   = 0x0000001A // Device is exclusive-access (R/W)
	SPDRP_CHARACTERISTICS             = 0x0000001B // Device Characteristics (R/W)
	SPDRP_ADDRESS                     = 0x0000001C // Device Address (R)
	SPDRP_UI_NUMBER_DESC_FORMAT       = 0x0000001D // UiNumberDescFormat (R/W)
	SPDRP_DEVICE_POWER_DATA           = 0x0000001E // Device Power Data (R)
	SPDRP_REMOVAL_POLICY              = 0x0000001F // Removal Policy (R)
	SPDRP_REMOVAL_POLICY_HW_DEFAULT   = 0x00000020 // Hardware Removal Policy (R)
	SPDRP_REMOVAL_POLICY_OVERRIDE     = 0x00000021 // Removal Policy Override (RW)
	SPDRP_INSTALL_STATE               = 0x00000022 // Device Install State (R)
	SPDRP_LOCATION_PATHS              = 0x00000023 // Device Location Paths (R)
	SPDRP_BASE_CONTAINERID            = 0x00000024 // Base ContainerID (R)
	SPDRP_MAXIMUM_PROPERTY            = 0x00000025 // Upper bound on ordinals
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
	devicePath string // e.g.: \\?\usb#vid_04f9&pid_2058#000f9z132168#{28d78fad-5a12-11d1-ae5b-0000f803a8c2}
	devInfo    spDevInfoData
}

func (hDevs *HDEVINFO) DevicePath() string {
	return hDevs.devicePath
}

func (hDevs *HDEVINFO) GetVidPid() (vid, pid uint16, err error) {
	// case A: \\?\usb#vid_6868&pid_0500&mi_00#6&29a28943&0&0000#{28d78fad-5a12-11d1-ae5b-0000f803a8c2}
	// case B: \\?\usb#vid_0fe6&pid_811e#38588749000935333146343453544632#{28d78fad-5a12-11d1-ae5b-0000f803a8c2}
	_, err = fmt.Sscanf(strings.ToLower(hDevs.devicePath), `\\?\usb#vid_%04x&pid_%04x`, &vid, &pid)
	return
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

func (hDevs *HDEVINFO) getDeviceRegistryProperty(property int) ([]uint16, error) {
	if hDevs.devInfo.DevInst == 0 {
		return nil, errors.New("invalid device")
	}
	dataType := 0
	var cbRequired uint32
	_, _, err := setupDiGetDeviceRegistryPropertyW.Call(uintptr(hDevs.h), uintptr(unsafe.Pointer(&hDevs.devInfo)), uintptr(property), uintptr(unsafe.Pointer(&dataType)), 0, 0, uintptr(unsafe.Pointer(&cbRequired)))
	if err != winspool.ERROR_INSUFFICIENT_BUFFER {
		return nil, err
	}
	pBuff := make([]uint16, cbRequired/2)
	r1, _, err := setupDiGetDeviceRegistryPropertyW.Call(uintptr(hDevs.h), uintptr(unsafe.Pointer(&hDevs.devInfo)), uintptr(property), uintptr(unsafe.Pointer(&dataType)), uintptr(unsafe.Pointer(&pBuff[0])), uintptr(cbRequired), uintptr(unsafe.Pointer(&cbRequired)))
	if r1 == 0 {
		return nil, err
	}
	return pBuff, nil
}

func (hDevs *HDEVINFO) GetLocation() (string, error) { // e.g.: Port_#0001.Hub_#0001
	buff, err := hDevs.getDeviceRegistryProperty(SPDRP_LOCATION_INFORMATION)
	if err != nil {
		return "", err
	}
	return syscall.UTF16ToString(buff), nil
}

func (hDevs *HDEVINFO) GetHardwareId() (string, error) { // e.g.: USB\VID_04F9&PID_2058&REV_0100
	buff, err := hDevs.getDeviceRegistryProperty(SPDRP_HARDWAREID)
	if err != nil {
		return "", err
	}
	return syscall.UTF16ToString(buff), nil
}

type HDevice syscall.Handle

func Open(devicePath string) (HDevice, error) {
	name, err := syscall.UTF16PtrFromString(devicePath)
	if err != nil {
		return 0, err
	}
	// https://docs.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-flushfilebuffers
	// https://docs.microsoft.com/en-us/windows/win32/fileio/file-buffering
	var dwFlagsAndAttributes uint32 = syscall.FILE_ATTRIBUTE_NORMAL | winapi.FILE_FLAG_NO_BUFFERING | winapi.FILE_FLAG_WRITE_THROUGH
	h, err := syscall.CreateFile(name, syscall.GENERIC_READ|syscall.GENERIC_WRITE, syscall.FILE_SHARE_READ, nil, syscall.OPEN_ALWAYS, dwFlagsAndAttributes, 0)
	return HDevice(h), err
}

func (hd HDevice) Close() error {
	// err := syscall.FlushFileBuffers(syscall.Handle(hd)) // get "Incorrect function." error
	return syscall.CloseHandle(syscall.Handle(hd))
}

func (hd HDevice) Write(data []byte) (int, error) {
	var written uint32
	err := syscall.WriteFile(syscall.Handle(hd), data, &written, nil)
	return int(written), err
}

func (hd HDevice) Read(p []byte) (int, error) {
	var done uint32
	err := syscall.ReadFile(syscall.Handle(hd), p, &done, nil)
	return int(done), err
}

func (hd HDevice) WriteAll(data []byte) (int, error) {
	total := 0
	for total < len(data) {
		n, err := hd.Write(data[total:])
		if err != nil {
			return total, err
		}
		if n <= 0 {
			return total, errors.New("can't write any more")
		}
		total += n
	}
	return total, nil
}

func (hd HDevice) ReadAll(p []byte) (int, error) {
	total := 0
	for total < len(p) {
		n, err := hd.Read(p[total:])
		if err != nil {
			return total, err
		}
		if n <= 0 {
			return total, errors.New("can't read any more")
		}
		total += n
	}
	return total, nil
}
