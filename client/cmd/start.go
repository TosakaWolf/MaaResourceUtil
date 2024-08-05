package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"maaResFetch/client/internal/config_client"
	"maaResFetch/common/dto"
	"maaResFetch/common/logger"
	"maaResFetch/common/utils"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	resourceDownload()
}

func resourceDownload() {
	// 创建或清空 tmp 目录
	tmpDir := "./tmpDownloadDir"
	err := os.RemoveAll(tmpDir)
	defer os.RemoveAll(tmpDir) // 解压完成后清理 tmp 文件夹
	if err != nil {
		fmt.Printf("移除临时下载目录失败: %v\n", err)
		return
	}
	err = os.MkdirAll(tmpDir, 0755)
	if err != nil {
		fmt.Printf("创建临时下载目录失败: %v\n", err)
		return
	}
	var downloadUrls []string
	downloadUrls = append(config_client.Config.ZipUrls, downloadUrls...)
	if len(config_client.Config.GetResourceUrls) > 0 {
		for _, url := range config_client.Config.GetResourceUrls {
			resourceUrl := getResourceUrl(url)
			if resourceUrl != "" {
				downloadUrls = append([]string{resourceUrl}, downloadUrls...)
			}
		}
	}
	// 下载并解压文件
	err = downloadAndExtractFromUrl(downloadUrls, tmpDir, config_client.Config.Directory)
	if err != nil {
		fmt.Printf("下载并解压资源文件失败: %v\n", err)
		return
	}

	fmt.Printf("已经下载并解压资源文件到: %s\n", config_client.Config.Directory)
}

func getResourceUrl(url string) string {
	request, err := http.NewRequest("GET", url, nil)
	// 发送 HTTP 请求
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		logger.Errorf("从服务端获取资源下载路径失败:%s", err.Error())
		return ""
	}
	defer resp.Body.Close()

	// Print the response.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var apiRes dto.ApiResult
	err = json.Unmarshal(body, &apiRes)
	if err != nil {
		return ""
	}
	return apiRes.Data.(string)
}

// downloadAndExtractFromUrl 下载并解压zip文件到指定目录
func downloadAndExtractFromUrl(urls []string, tmpDir string, outputDir string) error {
	var lastError error

	for _, url := range urls {
		logger.Infof("使用下载链接：%s" + url)
		zipFilePath := filepath.Join(tmpDir, "MaaResource-main.zip")

		err := utils.DownloadFile(url, zipFilePath)
		if err != nil {
			fmt.Printf("下载失败 %s: %v\n", url, err)
			lastError = err
			continue // 尝试下一个URL
		}

		// 解压下载的 zip 文件
		err = extractZip(zipFilePath, outputDir)
		if err != nil {
			fmt.Printf("解压失败 %s: %v\n", url, err)
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
