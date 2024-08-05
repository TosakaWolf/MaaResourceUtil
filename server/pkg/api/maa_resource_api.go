package api

import (
	"github.com/labstack/echo/v4"
	"maaResFetch/common/dto"
	"maaResFetch/server/pkg/service/file_service"
	"net/http"
)

func GetResource(c echo.Context) error {
	url := file_service.GetDownloadUrl()
	return c.JSON(http.StatusOK, dto.ApiResultSuccess(url))
}
