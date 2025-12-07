package common

import (
	"database/sql"
	"math/rand"
	"sync"
	"time"
)

var (
	Seed               = rand.New(rand.NewSource(time.Now().UnixNano()))
	Cfg                = Config{}
	Db         *sql.DB = nil
	Version            = "v1.4.3"
	Account            = 0
	CurrentBot int64   = 5
	Mutex      sync.Mutex
)

const ConfigDefaultFileName = "config.yaml"

type Config struct {
	Server struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"server"`
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Name     string `yaml:"name"`
	} `yaml:"database"`
	Log struct {
		Enabled  bool   `yaml:"enabled"`
		Output   string `yaml:"output"` // stderr, stdout, file, default is stdout
		Filepath string `yaml:"filepath"`
		Level    string `yaml:"level"` // Reserved
	} `yaml:"log"`
	Auth struct {
		JWTSecret string `yaml:"jwt_secret"`
	} `yaml:"auth"`
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

type ServerReply struct {
	Status  int               `json:"status"`
	Cmd     string            `json:"cmd"`
	TaskId  int64             `json:"taskid"`
	Args    map[string]any    `json:"args"`
	Error   string            `json:"error"`
	Headers map[string]string `json:"-"`
}
