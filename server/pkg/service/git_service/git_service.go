package git_service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"maaResourceUtil/common/logger"
	"maaResourceUtil/server/internal/config"
	"net/http"
	"os"
	"time"
)

const commitFileName = "latestCommit.json"

func CheckAndStoreLatestCommit() bool {
	request, err := http.NewRequest("GET", config.Config.CommitsUrl, nil)
	request.Header.Set("content-type", "application/json")
	// 发送 HTTP 请求
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		logger.Errorf("提交信息获取失败:%s", err.Error())
		return false
	}
	defer resp.Body.Close()
	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		logger.Errorf("Error: unexpected status code %d", resp.StatusCode)
		return false
	}

	// 解析响应体
	var commits []struct {
		Commit struct {
			Committer struct {
				Date time.Time `json:"date"`
			} `json:"committer"`
		} `json:"commit"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(body, &commits); err != nil {
		logger.Errorf("Error decoding JSON: %v ", err)
		return false
	}

	// 获取最新提交时间
	if len(commits) > 0 {
		latestCommit := commits[0]
		commitTime := latestCommit.Commit.Committer.Date
		// 检查是否发生变化
		if hasChanged(commitTime) {
			logger.Infof("最新提交时间: %v", commitTime.Format(time.RFC3339))
			logger.Info("提交时间初始化或者发生变化")
			// 如果发生了变化，存储最新提交时间
			storeLatestCommit(commitTime)
			return true
		} else {
			//logger.Info("提交时间未发生变化.")
		}
	} else {
		fmt.Println("未找到提交信息.")
	}
	return false
}

func hasChanged(newCommitTime time.Time) bool {
	// 读取之前存储的最新提交时间
	data, err := os.ReadFile(commitFileName)
	if err != nil {
		// 如果文件不存在或读取错误，认为发生了变化
		return true
	}

	// 解析 JSON 数据
	var storedData map[string]string
	if err := json.Unmarshal(data, &storedData); err != nil {
		log.Printf("Error decoding stored JSON: %v", err)
		return true
	}

	// 比较最新提交时间和之前存储的时间
	storedTimeStr, ok := storedData["latest_commit_time"]
	if !ok {
		log.Println("Stored data does not contain latest_commit_time.")
		return true
	}

	storedTime, err := time.Parse(time.RFC3339, storedTimeStr)
	if err != nil {
		log.Printf("Error parsing stored time: %v", err)
		return true
	}

	// 返回是否发生了变化
	return !newCommitTime.Equal(storedTime)
}

func storeLatestCommit(commitTime time.Time) {
	// 创建 JSON 对象
	data := map[string]interface{}{
		"latest_commit_time": commitTime.Format(time.RFC3339),
	}

	// 将 JSON 对象编码为 JSON 格式
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		logger.Errorf("Error encoding JSON: %v", err)
		return
	}

	// 将 JSON 数据写入文件
	err = os.WriteFile(commitFileName, jsonData, 0644)
	if err != nil {
		logger.Errorf("Error writing to file %s: %v", commitFileName, err)
		return
	}

	logger.Infof("最新提交时间信息保存到 %s", commitFileName)
}
