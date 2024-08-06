package config_client

import (
	"fmt"
	"maaResourceUtil/common/utils"
	"os"
)

// ClientConfig 结构用于存储目录配置信息
type ClientConfig struct {
	Directory       string   `yaml:"directory"` // maa目录路径
	GetResourceUrls []string `yaml:"getResourceUrls"`
	ZipUrls         []string `yaml:"zipUrls"` // 待下载的zip文件URL列表
}

var Config ClientConfig

func init() {
	// 配置文件路径
	configFilePath := "config-client.yaml"
	// 如果配置文件不存在，则创建一个默认的配置
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		defaultConfig := ClientConfig{
			Directory: "./maa_example", // 默认下载目录
			GetResourceUrls: []string{
				"http://127.0.0.1:8080/maa/getResource",
			},
			ZipUrls: []string{
				"https://github.com/MaaAssistantArknights/MaaResource/archive/refs/heads/main.zip",
			}, // 默认下载的 zip URLs
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
	Config = ClientConfig{}
	err := utils.ReadConfig(configFilePath, &Config)
	if err != nil {
		fmt.Printf("读取配置文件失败: %v\n", err)
		return
	}
}
