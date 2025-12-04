package components

import (
	"log"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"unsafe"

	"github.com/shirou/gopsutil/v4/cpu"
)

func get_os_version() string {
	return runtime.GOOS + "-" + runtime.GOARCH
}

func get_whoami() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "NaN"
	}
	username, err := user.Current()
	if err != nil {
		username.Username = "NaN"
	}
	return (hostname + "@" + username.Username)
}

func is_admin() bool {
	if runtime.GOOS == "windows" {
		var token uintptr
		currentProcess, _, _ := pfnGetCurrentProcess.Call()
		ok, _, _ := pfnOpenProcessToken.Call(currentProcess,
			TOKEN_QUERY, uintptr(unsafe.Pointer(&token)))
		if ok == 0 {
			log.Println("pfnOpenProcessToken failed: ")
			return false
		}
		defer pfnCloseHandle.Call(token)
		var tokenInfo uint32
		var outLen uint32
		ok, _, _ = pfnGetTokenInformation.Call(token, TokenElevation, uintptr(unsafe.Pointer(&tokenInfo)),
			uintptr(unsafe.Sizeof(tokenInfo)), uintptr(unsafe.Pointer(&outLen)))
		if ok == 0 {
			log.Println("pfnGetTokenInformation failed: ")
			return false
		}

		return uint32(tokenInfo) != 0
	} else {
		return os.Geteuid() == 0
	}
}

func get_cpu_info() string {
	info, err := cpu.Info()
	if err != nil || len(info) == 0 {
		return "Unknown CPU"
	}
	return info[0].ModelName
}

func get_gpu_info() string {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("powershell", "Get-CimInstance Win32_VideoController | Select-Object -ExpandProperty Name")
		byt, err := cmd.Output()
		if err != nil {
			return "Unknown GPU"
		}
		return string(byt)
	} else {
		cmd := exec.Command("lspci")
		byt, _ := cmd.Output()
		return string(byt)
	}
}
