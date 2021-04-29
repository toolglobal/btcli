package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
)

const (
	SUCCESS     = 0
	ParamError  = 1
	SysTemError = 500
)

type Resp struct {
	bodyWrite bool
}

type body struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
}

func (b *body) String() (str string, size int) {
	byt, _ := json.Marshal(b)
	return string(byt), len(byt)
}

func NewResp() *Resp {
	return &Resp{
		bodyWrite: true,
	}
}

func (res *Resp) RespOk(c *gin.Context) {
	res.JSON(200, body{
		Code:    0,
		Message: "success",
		Result:  "ok",
	}, c)
}

func (res *Resp) Resp(code int, msg string, c *gin.Context) {
	res.JSON(200, body{
		Code:    code,
		Message: msg,
	}, c)
}

func (res *Resp) RespError(code int, err error, c *gin.Context) {
	res.JSON(200, body{
		Code:    code,
		Message: err.Error(),
		Result:  "",
	}, c)
}

func (res *Resp) RespResult(result interface{}, c *gin.Context) {
	res.JSON(200, body{
		Code:    0,
		Message: "success",
		Result:  result,
	}, c)
}

func (res *Resp) JSON(httpCode int, b body, c *gin.Context) {
	if res.bodyWrite {
		str, size := b.String()
		if size <= (1024 << 10) {
			c.Set("resBody", str)
		}
	}
	c.JSON(httpCode, b)
}
