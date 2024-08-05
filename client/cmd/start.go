package main

import (
	"archive/zip"
	"fmt"
	"io"
	"maaResFetch/common/utils"
	"os"
	"path/filepath"
	"strings"
)

// Config 结构用于存储目录配置信息
type Config struct {
	Directory       string   `yaml:"directory"` // maa目录路径
	GetResourceUrls []string `yaml:"getResourceUrls"`
	ZipUrls         []string `yaml:"zipUrls"` // 待下载的zip文件URL列表
}

func main() {
	// 创建或清空 tmp 目录
	tmpDir := "./tmp"
	err := os.RemoveAll(tmpDir)
	defer os.RemoveAll(tmpDir) // 解压完成后清理 tmp 文件夹
	if err != nil {
		fmt.Printf("Failed to remove tmp directory: %v\n", err)
		return
	}
	err = os.MkdirAll(tmpDir, 0755)
	if err != nil {
		fmt.Printf("Failed to create tmp directory: %v\n", err)
		return
	}

	// 配置文件路径
	configFilePath := "config-client.yaml"

	// 如果配置文件不存在，则创建一个默认的配置
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		defaultConfig := Config{
			Directory: "./maa_example", // 默认下载目录
			ZipUrls: []string{
				"https://github.com/MaaAssistantArknights/MaaResource/archive/refs/heads/main.zip",
			}, // 默认下载的 zip URLs
		}
		err := utils.WriteConfig(configFilePath, defaultConfig)
		if err != nil {
			fmt.Printf("Failed to create config file: %v\n", err)
			return
		}
		fmt.Println("Created default config file:", configFilePath)
	}

	// 读取配置文件
	config := Config{}
	err = utils.ReadConfig(configFilePath, &config)
	if err != nil {
		fmt.Printf("Failed to read config file: %v\n", err)
		return
	}

	// 下载并解压文件
	err = downloadAndExtractFromUrls(config.ZipUrls, tmpDir, config.Directory)
	if err != nil {
		fmt.Printf("Failed to download and extract: %v\n", err)
		return
	}

	fmt.Printf("Downloaded and extracted to: %s\n", config.Directory)
}

// downloadAndExtractFromUrls 下载并解压来自多个URL的zip文件到指定目录
func downloadAndExtractFromUrls(urls []string, tmpDir string, outputDir string) error {
	var lastError error

	for _, url := range urls {
		zipFilePath := filepath.Join(tmpDir, "MaaResource-main.zip")

		err := utils.DownloadFile(url, zipFilePath)
		if err != nil {
			fmt.Printf("Failed to download from %s: %v\n", url, err)
			lastError = err
			continue // 尝试下一个URL
		}

		// 解压下载的 zip 文件
		err = extractZip(zipFilePath, outputDir)
		if err != nil {
			fmt.Printf("Failed to extract from %s: %v\n", url, err)
			lastError = err
			continue // 尝试下一个URL
		}

		// 下载和解压成功则退出循环
		lastError = nil
		break
	}

	return lastError
}

// extractZip 解压指定 zip 文件到目标目录
func extractZip(zipFilePath string, outputDir string) error {
	// 打开 zip 文件
	reader, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	// 创建目标目录
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		return err
	}

	// 解压文件
	for _, file := range reader.File {
		err := extractFileFromZip(file, outputDir)
		if err != nil {
			return err
		}
	}

	return nil
}

// extractFileFromZip 从 zip 中解压单个文件到目标目录
func extractFileFromZip(file *zip.File, outputDir string) error {
	// 打开文件
	reader, err := file.Open()
	if err != nil {
		return err
	}
	defer reader.Close()
	// 获取相对路径（去掉顶层文件夹路径）
	relativePath := strings.SplitN(file.Name, "/", 2)[1]

	// 创建目标文件
	extractedFilePath := filepath.Join(outputDir, relativePath)
	if file.FileInfo().IsDir() {
		os.MkdirAll(extractedFilePath, file.Mode())
	} else {
		outputFile, err := os.OpenFile(extractedFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer outputFile.Close()

		// 写入文件
		_, err = io.Copy(outputFile, reader)
		if err != nil {
			return err
		}
	}

	return nil
}
