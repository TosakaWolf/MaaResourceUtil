package api

import (
	"github.com/labstack/echo/v4"
	"maaResFetch/common/dto"
	"net/http"
)

func GetResource(c echo.Context) error {

	return c.JSON(http.StatusOK, dto.ApiResultSuccess(nil))
}
