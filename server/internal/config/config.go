package config

import (
	"fmt"
	"maaResFetch/common/utils"
	"os"
)

// ServerConfig 服务端配置文件
type ServerConfig struct {
	Port     int    `yaml:"port"` // maa目录路径
	ZipUrl   string `yaml:"zipUrl"`
	Cloud189 struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"cloud189"`
}

var Config ServerConfig

func init() {
	// 配置文件路径
	configFilePath := "config-server.yaml"

	// 如果配置文件不存在，则创建一个默认的配置
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		defaultConfig := ServerConfig{
			Port:   8080,
			ZipUrl: "https://github.com/MaaAssistantArknights/MaaResource/archive/refs/heads/main.zip",
		}
		err := utils.WriteConfig(configFilePath, defaultConfig)
		if err != nil {
			fmt.Printf("Failed to create config file: %v\n", err)
			return
		}
		fmt.Println("Created default config file:", configFilePath)
	}

	// 读取配置文件
	Config = ServerConfig{}
	err := utils.ReadConfig(configFilePath, &Config)
	if err != nil {
		fmt.Printf("Failed to read config file: %v\n", err)
		return
	}
}
