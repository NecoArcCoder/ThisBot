package components

import (
	"bytes"
	"log"
	"strings"
	"syscall"
	"unsafe"

	"github.com/StackExchange/wmi"
	"golang.org/x/sys/windows/registry"
)

func find_process_by_name(name string) bool {
	var dst []Win32_Process

	q := wmi.CreateQuery(&dst, "")
	err := wmi.Query(q, &dst)
	if err != nil {
		log.Fatal(err)
		return false
	}
	for _, v := range dst {
		if bytes.Contains([]byte(v.Name), []byte(name)) {
			return true
		}
	}
	return false
}

func find_pid_by_name(name string) uint32 {
	var pids [1024]uint32
	var bytesReturned uint32

	ret, _, _ := pfnEnumProcesses.Call(
		uintptr(unsafe.Pointer(&pids[0])),
		uintptr(len(pids))*unsafe.Sizeof(pids[0]),
		uintptr(unsafe.Pointer(&bytesReturned)),
	)
	if ret == 0 {
		return 0
	}

	count := uint32(bytesReturned / 4)

	for i := 0; i < int(count); i++ {
		pid := pids[i]
		if pid == 0 {
			continue
		}
		hProcess, _, _ := pfnOpenProcess.Call(PROCESS_QUERY_INFORMATION|PROCESS_VM_READ,
			0, uintptr(pid))
		if hProcess == 0 {
			continue
		}

		var hMod uintptr
		var cbNeeded uint32

		ret, _, _ = pfnEnumProcessModules.Call(hProcess,
			uintptr(unsafe.Pointer(&hMod)),
			unsafe.Sizeof(hMod),
			uintptr(unsafe.Pointer(&cbNeeded)),
		)
		if ret == 0 {
			pfnCloseHandle.Call(hProcess)
			continue
		}

		var exeName [260]uint16
		pfnGetModuleBaseNameW.Call(hProcess,
			hMod, uintptr(unsafe.Pointer(&exeName[0])), uintptr(len(exeName)))

		processName := syscall.UTF16ToString(exeName[:])

		if strings.EqualFold(strings.ToLower(processName), strings.ToLower(name)) {
			return pid
		}
	}

	return 0
}

func reg_read_key(key registry.Key, subPath string, value string, fInt bool) any {

	key1 := reg_create_key(key, subPath)
	if key1 == 0 {
		return nil
	}
	defer key1.Close()
	var data any

	if fInt {
		data, _, _ = key1.GetIntegerValue(value)
	} else {
		data, _, _ = key1.GetStringValue(value)
	}

	return data
}

func reg_delete_value(key registry.Key, subPath string, name string) bool {
	key1, err := registry.OpenKey(key, subPath, registry.ALL_ACCESS)
	if err != nil {
		return false
	}
	defer key1.Close()

	err = key1.DeleteValue(name)
	if err != nil {
		return false
	}
	return true
}

func reg_delete_key(key registry.Key, subPath string) bool {
	err := registry.DeleteKey(key, subPath)
	if err != nil {
		return false
	}
	return true
}

func reg_create_key(root registry.Key, subPath string) registry.Key {
	parts := strings.Split(subPath, `/`)

	k := root
	var err error
	for _, p := range parts {
		k, _, err = registry.CreateKey(k, p, registry.ALL_ACCESS)
		if err != nil {
			return 0
		}
	}
	return k
}

func reg_create_or_update_value(key registry.Key, subPath string, value string, data any, create bool) bool {
	var key1 registry.Key
	var err error

	if create {
		key1 = reg_create_key(key, subPath)
	} else {
		key1, err = registry.OpenKey(key, subPath, registry.ALL_ACCESS)
	}

	if err != nil {
		return false
	}
	defer key1.Close()

	switch v := data.(type) {
	case string:
		return key1.SetStringValue(value, v) == nil
	case uint32:
		return key1.SetDWordValue(value, v) == nil
	case uint64:
		return key1.SetQWordValue(value, v) == nil
	case []byte:
		return key1.SetBinaryValue(value, v) == nil
	case []string:
		return key1.SetStringsValue(value, v) == nil
	default:
		return false
	}
}
