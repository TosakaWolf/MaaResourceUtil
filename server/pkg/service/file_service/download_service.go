package file_service

import (
	"fmt"
	"maaResourceUtil/common/logger"
	"maaResourceUtil/server/internal/cloud_189"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	downloadUrlCache   string
	downloadUrlExpires time.Time
	cacheMutex         sync.Mutex
)

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
