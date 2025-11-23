package components

import (
	"syscall"
	"unsafe"
)

func create_mutex(name string) (uintptr, error) {
	wname, _ := syscall.UTF16PtrFromString(name)
	ret, _, err := pfnCreateMutexW.Call(0, 0, uintptr(unsafe.Pointer(wname)))
	switch int(err.(syscall.Errno)) {
	case 0:
		return ret, nil
	default:
		return ret, err
	}
}

func is_already_exist(name string) bool {
	_, err := create_mutex(name)
	if err != nil {
		return true
	} else {
		return false
	}
}
