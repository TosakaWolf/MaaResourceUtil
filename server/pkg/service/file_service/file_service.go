package file_service

import (
	"crypto/md5"
	"fmt"
	"github.com/tickstep/cloudpan189-api/cloudpan"
	"io"
	"maaResFetch/common/logger"
	"maaResFetch/common/utils"
	"maaResFetch/server/internal/cloud_189"
	"maaResFetch/server/internal/config"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var FileId string
var (
	downloadUrlCache   string
	downloadUrlExpires time.Time
	cacheMutex         sync.Mutex
)

func UploadResource() {
	panClient := cloud_189.GetPanClient()
	// 创建或清空 tmp 目录
	tmpDir := "./tmpDownloadDir"
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
	zipFilePath := filepath.Join(tmpDir, "MaaResource-main.zip")
	logger.Info("从github下载MaaResource资源文件")
	err = utils.DownloadFile(config.Config.ZipUrl, zipFilePath)
	if err != nil {
		fmt.Printf("Failed to download from %s: %v\n", config.Config.ZipUrl, err)
	}
	// 获取文件大小
	fileInfo, _ := os.Stat(zipFilePath)
	fileSize := fileInfo.Size()
	modTime := fileInfo.ModTime()

	// 格式化时间
	formattedTime := modTime.Format("2006-01-02 15:04:05")
	// 获取文件 MD5
	file, _ := os.Open(zipFilePath)
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		fmt.Println("计算 MD5 失败:", err)
	}
	md5sum := fmt.Sprintf("%x", hash.Sum(nil))
	var param cloudpan.AppCreateUploadFileParam
	param.LocalPath = zipFilePath
	param.FileName = filepath.Base(zipFilePath)
	param.ParentFolderId = "-11"
	param.Size = fileSize
	param.LastWrite = formattedTime
	param.Md5 = md5sum
	res, apiErr := panClient.AppCreateUploadFile(&param)
	if apiErr != nil {
		logger.Error(apiErr.Error())
		return
	}
	var fileRange cloudpan.AppFileUploadRange
	fileRange.Offset = 0
	fileRange.Len = fileSize
	logger.Info("上传MaaResource资源文件")
	panClient.AppUploadFileData(res.FileUploadUrl, res.UploadFileId, res.XRequestId, &fileRange,
		func(httpMethod, fullUrl string, headers map[string]string) (resp *http.Response, err error) {
			// 创建 HTTP 请求
			req, err := http.NewRequest(httpMethod, fullUrl, nil)
			if err != nil {
				return nil, err
			}

			// 设置 HTTP headers
			for key, value := range headers {
				req.Header.Set(key, value)
			}

			// 读取要上传的文件数据块
			file, err = os.Open(zipFilePath) // 替换为实际的文件路径
			if err != nil {
				return nil, err
			}
			defer file.Close()

			// 设置文件读取偏移量
			_, err = file.Seek(fileRange.Offset, io.SeekStart)
			if err != nil {
				return nil, err
			}

			// 创建读取器，限制读取长度
			reader := io.LimitReader(file, fileRange.Len)

			// 设置请求 body
			req.Body = io.NopCloser(reader)

			// 发送 HTTP 请求
			client := &http.Client{}
			resp, err = client.Do(req)
			if err != nil {
				return nil, err
			}

			// 返回响应结果
			return resp, nil
		})
	commitRes, commitErr := panClient.AppUploadFileCommitOverwrite(res.FileCommitUrl, res.UploadFileId, res.XRequestId, true)
	if commitErr != nil {
		logger.Errorf("文件上传失败：%s", commitErr.Error())
		return
	}
	FileId = commitRes.Id
	logger.Info("MaaResource资源文件上传完成")
}
func GetDownloadUrl() string {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if downloadUrlExpires.After(time.Now()) {
		return downloadUrlCache
	}

	panClient := cloud_189.GetPanClient()
	if FileId == "" {
		logger.Error("文件未上传，返回空")
		return ""
	}
	urlRes, urlerr := panClient.AppGetFileDownloadUrl(FileId)
	if urlerr != nil {
		logger.Error(urlerr.Error())
		return ""
	}

	expirationTime, err := parseExpirationTimeFromURL(urlRes)
	if err != nil {
		logger.Error(err.Error())
		return ""
	}
	logger.Infof("获取到新的下载链接：过期时间：%s", expirationTime.String())
	downloadUrlExpires = expirationTime.Add(-10 * time.Second)
	downloadUrlCache = urlRes

	return downloadUrlCache
}

func parseExpirationTimeFromURL(urlRes string) (time.Time, error) {
	// Example: https://download.cloud.189.cn/file/downloadFile.action?dt=n&expired=1722910619703&sk=xxx...
	expiredParam := "expired="
	expirationStart := strings.Index(urlRes, expiredParam)
	if expirationStart == -1 {
		return time.Time{}, fmt.Errorf("在url没有找到过期时间")
	}
	expirationStart += len(expiredParam)
	expirationEnd := strings.Index(urlRes[expirationStart:], "&")
	if expirationEnd == -1 {
		expirationEnd = len(urlRes)
	}
	expirationString := urlRes[expirationStart : expirationStart+expirationEnd]

	expirationUnix, err := strconv.ParseInt(expirationString, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("格式化过期时间错误: %v", err)
	}

	expirationTime := time.Unix(expirationUnix/1000, 0)
	return expirationTime, nil
}
