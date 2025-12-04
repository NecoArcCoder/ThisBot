package config

import (
	"ThisBot/common"
	"ThisBot/db1"
	"ThisBot/utils"
	"io"
	"log"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
)

// Generate random string with specific length
func GenerateRandom(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)
	for i := range bytes {
		bytes[i] = charset[common.Seed.Int()%len(charset)]
	}
	return string(bytes)
}

// Generate default configure for server and database(mysql)
func GenerateDefaultConfig(path string) error {
	cfg := common.Config{}

	cfg.Server.Host = "0.0.0.0"
	cfg.Server.Port = 8080

	cfg.Database.Host = "127.0.0.1"
	cfg.Database.Port = 3306
	cfg.Database.User = "master"
	cfg.Database.Password = "1234"
	cfg.Database.Name = "thisbot"

	cfg.Log.Enabled = false // Default no log
	cfg.Log.Filepath = "./Log/config.yaml"
	cfg.Log.Output = "stdout"
	cfg.Log.Level = "info"

	cfg.Auth.JWTSecret = GenerateRandom(48)

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		log.Print("yaml.Marshal failed")
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func LoadConfig(path string) *common.Config {
	var cfg common.Config

	data, _ := os.ReadFile(path)
	_ = yaml.Unmarshal(data, &cfg)

	return &cfg
}

func InitLog(cfg *common.Config) bool {
	if !cfg.Log.Enabled {
		// No log at all
		log.SetOutput(io.Discard)
		return true
	}

	log.SetPrefix("[ThisBot] ")
	log.SetFlags(log.Ldate | log.Lshortfile | log.Ltime)

	switch cfg.Log.Output {
	case "stdout":
		log.SetOutput(os.Stdout)
	case "stderr":
		log.SetOutput(os.Stderr)
	case "file":
		priv := os.O_WRONLY | os.O_APPEND

		logfile := "./Log"
		os.Mkdir(logfile, 0655)

		if !utils.FileExist(cfg.Log.Filepath) {
			priv |= os.O_CREATE
		}
		file, err := os.OpenFile(strings.Join([]string{logfile, common.ConfigDefaultFileName}, "/"), priv, 0644)
		if err != nil {
			log.Printf("os.OpenFile failed: %v\n", err)
			return false
		}
		log.SetOutput(file)
	default:
		log.SetOutput(os.Stdout)
	}
	return true
}

func Init(cfg *common.Config) bool {
	// Initialize log system
	if !InitLog(cfg) {
		log.Println("[-] Initialize log system failed")
	} else {
		log.Println("[+] Log system initialized successfully")
	}
	// Init mysql
	common.Db = db1.InitMysql(cfg.Database.User, cfg.Database.Password, cfg.Database.Name,
		cfg.Database.Host, cfg.Database.Port)
	if common.Db == nil {
		log.Fatal("db.InitMysql")
		return false
	}

	return true
}
