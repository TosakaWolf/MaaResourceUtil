package server

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"io"
	"maaResourceUtil/common/dto"
	"maaResourceUtil/common/logger"
	"maaResourceUtil/server/pkg/router"
	"net/http"
	"reflect"
	"time"
)

func New() (*echo.Echo, io.Closer) {
	e := echo.New()
	// Middleware
	var c io.Closer
	//trailing slash
	e.Pre(middleware.RemoveTrailingSlash())
	//cors
	e.Use(middleware.CORSWithConfig(middleware.DefaultCORSConfig))
	//secure
	e.Use(middleware.Secure())
	//api log
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		Skipper:   nil,
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger.Info("request",
				zap.String("URI", v.URI),
				zap.Int("status", v.Status),
			)
			return nil
		},
	}))

	//recover
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			if err != nil {
				logger.DPanic("recover err catch", zap.Error(err))
				fmt.Println(string(stack))
			}

			return err
		},
	}))

	//timeout
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		ErrorMessage: "time out",
		Timeout:      30 * time.Second,
	}))

	//rate limit
	//e.Pre(middleware.RateLimiterWithConfig(m2l_middleware.RateLimiterConfig()))

	//requestId [X-Request-ID]
	e.Use(middleware.RequestID())

	//bodyLimit
	e.Use(middleware.BodyLimit("5M"))

	//error handler
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if err != nil {
			// Send response
			var httpError *echo.HTTPError
			if reflect.TypeOf(err) == reflect.TypeOf(httpError) {
				errors.As(err, &httpError)
				err = c.JSON(httpError.Code, dto.ApiResultError(httpError.Message.(string)))
			} else {
				if c.Request().Method == http.MethodHead { // Echo Issue #608
					err = c.NoContent(http.StatusInternalServerError)
				} else {
					err = c.JSON(http.StatusInternalServerError, dto.ApiResultError(http.StatusText(http.StatusInternalServerError)))
				}
			}
		}
	}

	//接口初始化
	router.InitRouter(e)
	return e, c
}
