package m2l_middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"maaResourceUtil/common/dto"
	"maaResourceUtil/common/logger"
	"net/http"
	"time"
)

// RateLimiterConfig 限流中间件
func RateLimiterConfig() middleware.RateLimiterConfig {
	return middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			//每秒限制   并发阈值    清理冷却
			middleware.RateLimiterMemoryStoreConfig{Rate: 5, Burst: 15, ExpiresIn: 3 * time.Minute},
		),
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			id := ctx.RealIP()
			return id, nil
		},
		ErrorHandler: func(context echo.Context, err error) error {
			logger.Error("Rate Limit Error", zap.String("ERROR", err.Error()))
			return context.JSON(http.StatusForbidden, dto.ApiResultError("T^T"))
		},
		DenyHandler: func(context echo.Context, identifier string, err error) error {
			logger.Error("Rate Limited", zap.String("IP", context.RealIP()))
			return context.JSON(http.StatusTooManyRequests, dto.ApiResultError("QAQ"))
		},
	}
}
