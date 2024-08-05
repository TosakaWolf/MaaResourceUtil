package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// DownloadFile 下载文件到指定路径
func DownloadFile(url string, filePath string) error {
	// 创建目标文件
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// 发起 GET 请求
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file, status code: %d", resp.StatusCode)
	}

	// 将响应内容写入文件
	_, err = io.Copy(out, resp.Body)
	return err
}
