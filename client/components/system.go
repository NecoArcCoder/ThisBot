package components

import (
	"bytes"
	"log"
	"strings"
	"syscall"
	"unsafe"

	"github.com/StackExchange/wmi"
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
