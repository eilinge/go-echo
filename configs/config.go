package configs

import (
	"flag"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Version ...
var (
	Version   = "1.0.0"
	Commit    = ""
	BuildTime = "2018-10-09"
)

// ServerConfig ...
type ServerConfig struct {
	Common *CommonConfig
	Db     *DbConfig
	Eth    *EthConfig
}

// CommonConfig ...
type CommonConfig struct {
	Port      string
	LogFormat string
}

// DbConfig  ...
type DbConfig struct {
	Driver  string
	Connstr string
}

// EthConfig ...
type EthConfig struct {
	Connstr      string
	Keydir       string
	Fundation    string
	PxcAddr      string
	PxaAddr      string
	FundationPWD string
}

func usage() {
	fmt.Printf("Usage: %s -c config_file [-v] [-h]\n", os.Args[0])
}

// Config ...
var Config *ServerConfig //引用配置文件结构

func init() {
	fmt.Println("call config.init")
	Config = GetConfig()
}

// GetConfig ...
func GetConfig() (config *ServerConfig) {
	var configFile = flag.String("c", "", "Config file")

	var ver = flag.Bool("v", false, "version")
	var help = flag.Bool("h", false, "Help")

	flag.Usage = usage
	flag.Parse()

	if *help {
		usage()
		return nil
	}

	if *ver {
		fmt.Println("Version: ", Version)
		fmt.Println("Commit: ", Commit)
		fmt.Println("BuildTime: ", BuildTime)
		return nil
	}
	// get server config
	if *configFile != "" {
		config = &ServerConfig{}
		// 文件提取
		if _, err := toml.DecodeFile(*configFile, &config); err != nil {
			panic(err)
		}
	} else {
		usage()
		return nil
	}

	return config
}
