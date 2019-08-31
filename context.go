package xgin

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	data_kind_json = iota + 1
	data_kind_string
	data_kind_msg_data
)

type MsgData struct {
	Data    interface{} `json:"data"`
	Msg     string      `json:"msg"`
	Errcode int         `json:"errcode"`
}

// gin框架上下文的扩展
type XGin struct {
	*gin.Context
	statusCode int         //响应码
	data       interface{} //响应数据
	msgData    *MsgData    //对响应数据的扩展
	dataKind   int         //响应数据类型(json,string,...)
}

func Get(ctx *gin.Context) *XGin {
	xgin, ok := ctx.Get("xgin")
	if ok {
		return xgin.(*XGin)
	}
	xgin = &XGin{
		Context:    ctx,
		statusCode: http.StatusOK,  //默认响应码
		dataKind:   data_kind_json, //默认写数据格式
	}
	ctx.Set("xgin", xgin)
	return xgin.(*XGin)
}

func (xg *XGin) GetMsgData() *MsgData {
	if xg.msgData == nil {
		xg.dataKind = data_kind_msg_data
		xg.msgData = &MsgData{
			Errcode: 1,    //默认错误码
			Msg:     "ok", //默认消息
		}
	}
	return xg.msgData
}

func (xg *XGin) StatusCode(statusCode int) *XGin {
	xg.statusCode = statusCode
	return xg
}

func (xg *XGin) Json(data interface{}) *XGin {
	xg.data = data
	xg.dataKind = data_kind_json
	return xg
}

func (xg *XGin) String(str string) *XGin {
	xg.data = str
	xg.dataKind = data_kind_string
	return xg
}

func (xg *XGin) MsgData(data interface{}) *XGin {
	xg.GetMsgData().Data = data
	return xg
}

func (xg *XGin) Errcode(errcode int) *XGin {
	xg.GetMsgData().Errcode = errcode
	return xg
}

func (xg *XGin) Msg(msg string) *XGin {
	xg.GetMsgData().Msg = msg
	return xg
}

func (xg *XGin) Write() {
	switch xg.dataKind {
	case data_kind_json:
		xg.Context.JSON(xg.statusCode, xg.data)
	case data_kind_string:
		xg.Context.String(xg.statusCode, xg.data.(string))
	case data_kind_msg_data:
		xg.Context.JSON(xg.statusCode, xg.msgData)
	default:
		panic(fmt.Sprintf("不支持的数据响应格式类型:%d", xg.dataKind))
	}
}

// 配置数据库连接接口
func (xg *XGin) IniConScope() {
	xg.Context.Set("conscope", "TODO")
	log.Println("xgin iniConScope TODO")
}

func (xg *XGin) ConScope() interface{} {
	return xg.Context.MustGet("conscope")
}
