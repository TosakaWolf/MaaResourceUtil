package utils

import (
	"gopkg.in/yaml.v3"
	"os"
)

// WriteConfig 将配置写入 YAML 文件
func WriteConfig(filePath string, config any) error {

	// 创建或更新文件
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 编码为 YAML 并写入文件
	encoder := yaml.NewEncoder(file)
	defer encoder.Close()

	// 设置 yaml 格式
	encoder.SetIndent(2)

	// 将带注释的配置结构体写入文件
	err = encoder.Encode(config)
	if err != nil {
		return err
	}

	return nil
}

// ReadConfig 读取配置文件
func ReadConfig(filePath string, config any) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		return err
	}

	return nil
}
