package apiresp

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"lbe_crypto_signer/pkg/common/errs"
	"lbe_crypto_signer/pkg/util/jsonutil"
	"net/http"
	"reflect"
	"strconv"
)

type ApiResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Dlt  string `json:"dlt"`
	Data any    `json:"data,omitempty"`
}

func (r *ApiResponse) MarshalJSON() ([]byte, error) {
	type apiResponse ApiResponse
	tmp := (*apiResponse)(r)
	if tmp.Data != nil {

		if isAllFieldsPrivate(tmp.Data) {
			tmp.Data = json.RawMessage(nil)
		} else {
			data, err := jsonutil.JsonMarshal(tmp.Data)
			if err != nil {
				return nil, err
			}
			tmp.Data = json.RawMessage(data)
		}
	}
	return jsonutil.JsonMarshal(tmp)
}

func isAllFieldsPrivate(v any) bool {
	typeOf := reflect.TypeOf(v)
	if typeOf == nil {
		return false
	}
	for typeOf.Kind() == reflect.Ptr {
		typeOf = typeOf.Elem()
	}
	if typeOf.Kind() != reflect.Struct {
		return false
	}
	num := typeOf.NumField()
	for i := 0; i < num; i++ {
		c := typeOf.Field(i).Name[0]
		if c >= 'A' && c <= 'Z' {
			return false
		}
	}
	return true
}

func apiSuccess(data any) *ApiResponse {
	return &ApiResponse{Data: data}
}

func ginJson(c *gin.Context, resp *ApiResponse) {
	if resp.Code == 0 {
		c.Set("result", "success")
	} else {
		c.Set("result", "failure")
		c.Set("error_code", strconv.Itoa(resp.Code))
	}
	c.JSON(http.StatusOK, resp)
}

func ApiError(c *gin.Context, err error) {
	ginJson(c, ParseError(err))
}

func ApiSuccess(c *gin.Context, data any) {
	ginJson(c, apiSuccess(data))
}

func ParseError(err error) *ApiResponse {
	if err == nil {
		return apiSuccess(nil)
	}

	unwrap := errs.Unwrap(err)
	if codeErr, ok := unwrap.(errs.CodeError); ok {
		resp := ApiResponse{Code: codeErr.Code(), Msg: codeErr.Msg(), Dlt: codeErr.Detail()}
		if resp.Dlt == "" {
			resp.Dlt = err.Error()
		}
		return &resp
	}
	return &ApiResponse{Code: errs.ServerInternalError, Msg: err.Error()}
}
