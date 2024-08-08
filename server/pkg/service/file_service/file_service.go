package file_service

import (
	"crypto/md5"
	"fmt"
	"github.com/tickstep/cloudpan189-api/cloudpan"
	"io"
	"maaResourceUtil/common/logger"
	"maaResourceUtil/common/utils"
	"maaResourceUtil/server/internal/cloud_189"
	"maaResourceUtil/server/internal/config"
	"maaResourceUtil/server/pkg/service/git_service"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var FileId string

var (
	uploadQueue chan struct{}
	uploadMutex sync.Mutex
)
var fileIdFileName = "fileId.json"

func init() {
	uploadQueue = make(chan struct{}, 1) // 使用缓冲大小为1的通道作为任务队列
	// 启动任务处理
	go processUploadQueue()
	// 程序启动时从文件中加载 FileId
	loadFileIdFromFile()

	if FileId == "" {
		uploadQueue <- struct{}{}
	}

	// 每隔十分钟检查一次
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				changed := git_service.CheckAndStoreLatestCommit()
				if changed {
					// 加入任务队列
					uploadQueue <- struct{}{}
				}
			}
		}
	}()

}

func processUploadQueue() {
	for range uploadQueue {
		// 加锁等待资源更新完成
		uploadMutex.Lock()
		UploadResource()
		uploadMutex.Unlock()
	}
}

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

	// 保存 FileId 到文件中
	saveFileIdToFile()
}

func saveFileIdToFile() {
	file, err := os.Create(fileIdFileName)
	if err != nil {
		logger.Errorf("无法创建文件 %s: %v", fileIdFileName, err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(FileId)
	if err != nil {
		logger.Errorf("无法写入文件 %s: %v", fileIdFileName, err)
		return
	}

	logger.Infof("FileId 已保存到文件 %s", fileIdFileName)
}

func loadFileIdFromFile() {
	// 尝试从文件中加载 FileId
	file, err := os.Open(fileIdFileName)
	if err != nil {
		//logger.Warnf("文件不存在 %s: %v", fileIdFileName, err)
		return
	}
	defer file.Close()

	// 读取 FileId
	var id string
	_, err = fmt.Fscanf(file, "%s", &id)
	if err != nil {
		logger.Errorf("无法从文件 %s 中读取 FileId: %v", fileIdFileName, err)
		return
	}

	FileId = id
	logger.Infof("成功从文件 %s 中加载 FileId: %s", fileIdFileName, FileId)
}
