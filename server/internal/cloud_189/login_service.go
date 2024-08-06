package cloud_189

import (
	"encoding/json"
	"os"
	"time"

	"github.com/tickstep/cloudpan189-api/cloudpan"
	"go.uber.org/zap"
	"maaResFetch/common/logger"
	"maaResFetch/server/internal/config"
)

var AppToken cloudpan.AppLoginToken
var appTokenFlag bool
var PanClient *cloudpan.PanClient

const appTokenFile = "appToken.json"
const getClientLock = "lock:189:getPanClient"

func init() {
	go panClientResetTicker()
}

func panClientResetTicker() {
	ticker := time.NewTicker(24 * time.Hour)
	for {
		select {
		case <-ticker.C:
			logger.Info("24h 自动清理189 PanClient")
			PanClient = nil
			appTokenFlag = false
			GetPanClient()
		}
	}
}

func GetAppToken(username, password string) {
	appTokenFlag = true

	// 检查 appToken.json 文件的创建时间
	if tokenFromFile, ok := readTokenFromFile(); ok {
		AppToken = *tokenFromFile
		logger.Info("使用appToken.json登录")
		return
	}

	// 如果 appToken.json 文件超过三天创建，重新生成 token 并写入 appToken.json
	token, apiErr := cloudpan.AppLogin(username, password)
	if apiErr != nil {
		logger.Error("189 login fail", zap.Error(apiErr))
		appTokenFlag = false
		return
	}
	AppToken = *token

	// 将 token 写入 appToken.json
	appTokenJson, err := json.Marshal(token)
	if err != nil {
		logger.Error("189 appToken jsonMarshal Err", zap.Error(err))
		appTokenFlag = false
		return
	}
	err = os.WriteFile(appTokenFile, appTokenJson, 0755)
	if err != nil {
		logger.Error("Failed to write appToken.json", zap.Error(err))
		appTokenFlag = false
	} else {
		logger.Info("appToken.json written successfully")
	}
}

func readTokenFromFile() (*cloudpan.AppLoginToken, bool) {
	fileInfo, err := os.Stat(appTokenFile)
	if err != nil {
		logger.Info("首次登录，将创建appToken.json")
		return nil, false
	}
	fileAge := time.Since(fileInfo.ModTime()).Hours() / 24

	if fileAge <= 3 {
		fileData, err := os.ReadFile(appTokenFile)
		if err != nil {
			logger.Error("Failed to read appToken.json", zap.Error(err))
			return nil, false
		}
		var token cloudpan.AppLoginToken
		err = json.Unmarshal(fileData, &token)
		if err != nil {
			logger.Error("Failed to unmarshal appToken.json", zap.Error(err))
			return nil, false
		}
		return &token, true
	}
	return nil, false
}

func Login() *cloudpan.PanClient {
	if !appTokenFlag {
		logger.Error("189 appTokenFlag false")
		return nil
	}
	webToken := &cloudpan.WebLoginToken{}
	webTokenStr := GetWebTokenStr()
	if webTokenStr != "" {
		webToken.CookieLoginUser = webTokenStr
	} else {
		logger.Error("webToken获取失败，如果多次获取失败请手动清理appToken.json.如果切换ip请直接删除")
	}
	// pan client
	panClient := cloudpan.NewPanClient(*webToken, AppToken)
	info, apiErr := panClient.GetUserInfo()
	if apiErr != nil {
		logger.Error("189 login error", zap.Error(apiErr))
		return nil
	}
	logger.Info("189 login success")
	logger.Info("189 GetUserInfo", zap.Any("UserInfo", info))
	return panClient
}

func GetWebTokenStr() string {
	webTokenStr := cloudpan.RefreshCookieToken(AppToken.SessionKey)
	return webTokenStr
}
func GetPanClient() *cloudpan.PanClient {
	if PanClient == nil {
		logger.Info("初始化189网盘")
		tryTimes := 0
		initializeClient(&tryTimes)
	}
	return PanClient
}

func initializeClient(tryTimes *int) {
	*tryTimes = *tryTimes + 1
	logger.Info("189 client 初始化", zap.Int("尝试次数", *tryTimes))
	GetAppToken(config.Config.Cloud189.Username, config.Config.Cloud189.Password)
	PanClient = Login()
	if PanClient == nil {
		for *tryTimes < 3 {
			initializeClient(tryTimes)
		}
		if *tryTimes == 3 {
			logger.Panic("189 client 初始化失败")
		}
	}
}
