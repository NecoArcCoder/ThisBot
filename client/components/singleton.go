package components

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func is_already_exist(name string) (bool, uintptr) {
	if runtime.GOOS == "windows" {
		wname, _ := syscall.UTF16PtrFromString(name)
		fmt.Println("mutex: ", name, ", pid: ", os.Getpid())
		ret, _, _ := pfnCreateMutexW.Call(0, 0, uintptr(unsafe.Pointer(wname)))
		if ret != 0 {
			code, _, _ := pfnGetLastError.Call()
			if code == uintptr(windows.ERROR_ALREADY_EXISTS) {
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
