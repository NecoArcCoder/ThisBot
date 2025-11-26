package components

func is_debugger_exist() bool {
	ret, _, _ := pfnIsDebuggerPresent.Call()
	if ret != 0 {
		return true
	}

	for i := 0; i < len(debuggers); i++ {
		if find_process_by_name(debuggers[i]) {
			return true
		}
	}

	return false
}

func in_sandbox_now() bool {
	return false
}

func in_vm_now() bool {
	return false
}

func install_payload() {

}
