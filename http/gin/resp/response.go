package resp

import (
	"encoding/json"
	"net/http"

	xgin "miopkg/http/gin"

	"github.com/gin-gonic/gin"
)

type RespPagination struct {
	List       interface{} `json:"list,omitempty"`
	Total      int32       `json:"total,omitempty"`
	Pagination struct {
		Page int32 `json:"offset,omitempty"`
		Size int32 `json:"limit,omitempty"`
	} `json:"pagination,omitempty"`
}
type RespCursor struct {
	List   interface{} `json:"list,omitempty"`
	Total  int32       `json:"total,omitempty"`
	Cursor struct {
		Offset int32 `json:"offset,omitempty"`
		Limit  int32 `json:"limit,omitempty"`
	} `json:"cursor,omitempty"`
}

// JSONResult json
type JSONResult struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type JSONResultRaw struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// JSON 提供了系统标准JSON输出方法。
func JSON(c *gin.Context, Code int, data ...interface{}) {
	result := new(JSONResult)
	result.Code = Code
	info := xgin.StatusText(Code)
	result.Message = info

	if len(data) > 0 {
		result.Data = data[0]
	} else {
		result.Data = ""
	}
	c.JSON(http.StatusOK, result)
	return
}

func JSONSuccess(c *gin.Context, data ...interface{}) {
	result := new(JSONResult)
	result.Code = 0
	result.Message = "success"
	if len(data) > 0 {
		result.Data = data[0]
	} else {
		result.Data = ""
	}
	c.JSON(http.StatusOK, result)
	return
}

// JSON 提供了系统标准JSON输出方法
func JSONErr(c *gin.Context, Code int) {
	result := new(JSONResult)
	result.Code = Code
	msg := xgin.StatusText(Code)
	result.Message = msg
	c.JSON(http.StatusOK, result)
	return
}

// func JSONErr(c *gin.Context, Code int, err error) {
// 	result := new(JSONResult)
// 	result.Code = Code
// 	info := xgin.StatusText(Code)
// 	result.Message = info
// 	if err != nil {
// 		fmt.Println("code is", Code, "info is", result.Message, "err is", err.Error())
// 	}
// 	c.JSON(http.StatusOK, result)
// 	return
// }

func JSONPagination(c *gin.Context, data interface{}, total, page, size int32) {
	j := new(JSONResult)
	j.Code = 0
	j.Message = "success"
	j.Data = RespPagination{
		List:  data,
		Total: total,
		Pagination: struct {
			Page int32 `json:"offset,omitempty"`
			Size int32 `json:"limit,omitempty"`
		}{
			Page: page,
			Size: size,
		},
	}
	c.JSON(http.StatusOK, j)
	return
}
func JSONCursor(c *gin.Context, data interface{}, total, offset, limit int32) {
	j := new(JSONResult)
	j.Code = 0
	j.Message = "success"
	j.Data = RespCursor{
		List:  data,
		Total: total,
		Cursor: struct {
			Offset int32 `json:"offset,omitempty"`
			Limit  int32 `json:"limit,omitempty"`
		}{
			Offset: offset,
			Limit:  limit,
		},
	}
	c.JSON(http.StatusOK, j)
	return
}
