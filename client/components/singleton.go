package components

import (
	"errors"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func is_already_exist(name string) (bool, uintptr) {
	if runtime.GOOS == "windows" {
		mutex_name := "Global\\" + name
		wname, _ := syscall.UTF16PtrFromString(mutex_name)
		ret, _, err := pfnCreateMutexW.Call(0, 0, uintptr(unsafe.Pointer(wname)))
		if ret != 0 {
			if errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
				pfnCloseHandle.Call(ret)
				return true, 0
			}
		}
		return false, ret
	} else {
		// linux
		return false, 0
	}
}
