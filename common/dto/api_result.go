package dto

import "net/http"

type ApiResult struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Success bool        `json:"success"`
}

func ApiResultSuccess(data interface{}) ApiResult {
	return ApiResult{
		Code:    http.StatusOK,
		Data:    data,
		Message: "操作成功",
		Success: true,
	}
}

func ApiResultSuccessWithMessage(data interface{}, message string) ApiResult {
	return ApiResult{
		Code:    http.StatusOK,
		Data:    data,
		Message: message,
		Success: true,
	}
}

func ApiResultErrorWithCode(code int, message string) ApiResult {
	return ApiResult{
		Code:    code,
		Data:    nil,
		Message: message,
		Success: false,
	}
}

func ApiResultError(message string) ApiResult {
	return ApiResult{
		Code:    http.StatusInternalServerError,
		Data:    nil,
		Message: message,
		Success: false,
	}
}
