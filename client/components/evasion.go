package components

import (
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
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
		cmd := exec.Command("powershell", "-Command", "Get-CimInstance Win32_ComputerSystem | Select-Object Manufacturer, Model")
		out, err := cmd.Output()
		if err != nil {
			return false
		}
		lower := strings.ToLower(string(out))
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

}
