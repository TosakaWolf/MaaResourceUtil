package config

import (
	"fmt"
	"maaResourceUtil/common/utils"
	"os"
)

// ServerConfig 服务端配置文件
type ServerConfig struct {
	Port       int    `yaml:"port"` // maa目录路径
	ZipUrl     string `yaml:"zipUrl"`
	CommitsUrl string `yaml:"commitsUrl"`
	Cloud189   struct {
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
			Port:       8080,
			ZipUrl:     "https://github.com/MaaAssistantArknights/MaaResource/archive/refs/heads/main.zip",
			CommitsUrl: "https://github.com/MaaAssistantArknights/MaaResource/commits",
		}
		err := utils.WriteConfig(configFilePath, defaultConfig)
		if err != nil {
			fmt.Printf("创建配置文件失败: %v\n", err)
			return
		}
		fmt.Println("已经自动创建配置文件:", configFilePath)
		fmt.Println("请修改配置文件")
		os.Exit(1)
	}

	// 读取配置文件
	Config = ServerConfig{}
	err := utils.ReadConfig(configFilePath, &Config)
	if err != nil {
		fmt.Printf("Failed to read config file: %v\n", err)
		return
	}
}
