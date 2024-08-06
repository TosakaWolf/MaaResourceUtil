package cloud_189

import (
	"github.com/tickstep/cloudpan189-api/cloudpan"
	"go.uber.org/zap"
	"maaResFetch/common/logger"
	"maaResFetch/server/internal/config"
	"time"
)

var AppToken cloudpan.AppLoginToken
var appTokenFlag bool
var PanClient *cloudpan.PanClient

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

func GetAppToken(username string, password string) {
	appTokenFlag = true

	token, apiErr := cloudpan.AppLogin(username, password)
	if apiErr != nil {
		logger.Error("189 login fail", zap.Error(apiErr))
		appTokenFlag = false
	}
	AppToken = *token
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
		logger.Error("webToken为空")
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
		if PanClient == nil {
			tryTimes := 0
			initializeClient(&tryTimes)
		}
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
