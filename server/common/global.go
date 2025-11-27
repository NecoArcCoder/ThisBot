package common

import (
	"database/sql"
	"math/rand"
	"time"
)

var (
	Seed         = rand.New(rand.NewSource(time.Now().UnixNano()))
	Cfg          = Config{}
	Db   *sql.DB = nil
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

// type Account struct {
// 	Id       int
// 	Username string
// 	Password string
// }

type Client struct {
	Id          int
	Guid        string
	Token       string
	Ip          string
	Whoami      string
	Os          string
	Installdate string
	Isadmin     string
	Antivirus   string
	Cpuinfo     string
	Gpuinfo     string
	Version     string
	Lastseen    string
	Lastcommand string
}

// type Command struct {
// 	Id          int
// 	Command     string
// 	Timeanddate string
// }

// type Lastlogin struct {
// 	Id          int
// 	Timeanddate string
// }

// type Tasks struct {
// 	Id      int
// 	Name    string
// 	Guid    string
// 	Command string
// 	Method  string
// }

type ServerReply struct {
	Status  int               `json:"status"`
	Cmd     string            `json:"cmd"`
	TaskId  int64             `json:"taskid"`
	Args    map[string]any    `json:"args"`
	Error   string            `json:"error"`
	Headers map[string]string `json:"-"`
}
