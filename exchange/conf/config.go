package conf

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/bCoder778/log"
	"os"
	"sync"
)

const (
	configFile = "config.toml"
)

var Setting *Config
var once sync.Once

func init() {
	once.Do(func() {
		if Exist(configFile) {
			if _, err := toml.DecodeFile(configFile, &Setting); err != nil {
				fmt.Println(err)
			}
		}
	})
}

type Config struct {
	Api  *Api  `toml:"api"`
	Rpc  *Rpc  `toml:"rpc"`
	Sync *Sync `toml:"sync"`
	Log *Log `toml:"log"`
}

type Api struct {
	Listen string `toml:"listen"`
}

type Rpc struct {
	Host     string `toml:"host"`
	Tls      bool   `toml:"tls"`
	Admin    string `toml:"admin"`
	Password string `toml:"password"`
}

type Sync struct {
	Start         uint64   `toml:"start"`
	Confirmations uint64   `toml:"confirmations"`
	Address       []string `toml:"address"`
	Log           *Log     `toml:"log"`
}

type Log struct {
	Mode  log.Mode  `toml:"mode"`
	Level log.Level `toml:"level"`
	Path  string    `toml:"path"`
}

func Exist(fileName string) bool {
	_, err := os.Stat(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
