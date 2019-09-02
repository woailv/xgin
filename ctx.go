package xgin

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tealeg/xlsx"
)

const (
	data_kind_json = iota + 1
	data_kind_string
	data_kind_msg_data
)

const (
	query_key_page     = "page"
	query_key_item     = "item"
	query_key_download = "download"
	query_key_sort     = "sort"
)

type MsgData struct {
	Data    interface{} `json:"data"`
	Msg     string      `json:"msg"`
	Errcode int         `json:"code"`

	sort string //Data为Slice时的排序字段
}

type ListData struct {
	List  interface{} `json:"data"`
	Total int         `json:"total"`
}

// gin框架上下文的扩展
type XCtx struct {
	*gin.Context
	statusCode int         //响应码
	data       interface{} //响应数据
	msgData    *MsgData    //对响应数据的扩展
	dataKind   int         //响应数据类型(json,string,...)
}

func Get(ctx *gin.Context) *XCtx {
	xCtx, ok := ctx.Get("xCtx")
	if ok {
		return xCtx.(*XCtx)
	}
	xCtx = &XCtx{
		Context:    ctx,
		statusCode: http.StatusOK,  //默认响应码
		dataKind:   data_kind_json, //默认写数据格式
	}
	ctx.Set("xCtx", xCtx)
	return xCtx.(*XCtx)
}

func (xc *XCtx) GetMsgData() *MsgData {
	if xc.msgData == nil {
		xc.dataKind = data_kind_msg_data
		xc.msgData = &MsgData{
			Errcode: 1,    //默认错误码
			Msg:     "ok", //默认消息
		}
	}
	return xc.msgData
}

func (xc *XCtx) StatusCode(statusCode int) *XCtx {
	xc.statusCode = statusCode
	return xc
}

func (xc *XCtx) Json(data interface{}) *XCtx {
	// 数值类型保留两位小数
	xc.data = getDecimal2(data)
	xc.dataKind = data_kind_json
	return xc
}

func (xc *XCtx) String(str string) *XCtx {
	xc.data = str
	xc.dataKind = data_kind_string
	return xc
}

func (xc *XCtx) MsgData(data interface{}) *XCtx {
	// 数值类型保留两位小数
	xc.GetMsgData().Data = getDecimal2(data)
	return xc
}

func (xc *XCtx) Errcode(errcode int) *XCtx {
	xc.GetMsgData().Errcode = errcode
	return xc
}

func (xc *XCtx) Msg(msg string) *XCtx {
	xc.GetMsgData().Msg = msg
	return xc
}

// 响应错误并将errcode设置为2
func (xc *XCtx) Err(err error) *XCtx {
	md := xc.GetMsgData()
	md.Msg = err.Error()
	md.Errcode = 2
	return xc
}

func assert(err error) {
	if err != nil {
		panic(err)
	}
}

func (xc *XCtx) QueryInt(key string) int {
	val := xc.Query(key)
	if val == "" {
		val = "0"
	}
	v, err := strconv.Atoi(val)
	assert(err)
	return v
}

// 获取分页数据的页码
func (xc *XCtx) Page() int {
	return xc.QueryInt(query_key_page)
}

// 获取分页数据的每页条数
func (xc *XCtx) Item() int {
	v := xc.QueryInt(query_key_item)
	if v == 0 {
		v = 10
	}
	return v
}

// 响应分页数据
func (xc *XCtx) List(list interface{}, total int) *XCtx {
	// 数值类型保留两位小数
	listData := &ListData{List: getDecimal2(list), Total: total}
	xc.GetMsgData().Data = listData
	return xc
}

func skip(page, item int) int {
	if page == 0 {
		return 0
	}
	return (page - 1) * limit(page, item)
}

func limit(page, item int) int {
	if item == 0 && page != 0 {
		return 10
	} else if item != 0 && page != 0 {
		return item
	}
	return 0
}

// 返回列表中的默认排序字段
func (xc *XCtx) sortDef(str string) *XCtx {
	md := xc.GetMsgData()
	if s := xc.Query(query_key_sort); s == "" {
		md.sort = str
	} else {
		md.sort = s
	}
	log.Println("排序字段:", md.sort)
	return xc
}

// 接受客户端排序字段
func (xc *XCtx) sort() *XCtx {
	return xc.sortDef("")
}

// 响应对切片分页的数据(有分页操作)
// 有排序字段要先调用排序字段的方法
func (xc *XCtx) Slice(slice interface{}) *XCtx {
	list, total := sliceData(slice, xc.Page(), xc.Item())
	return xc.List(list, total)
}

func (xc *XCtx) SliceSortDef(slice interface{}, defSort string) *XCtx {
	xc.sortDef(defSort)
	if sort := xc.GetMsgData().sort; sort != "" { //排序
		slice = getSortSlice(slice, sort)
	}
	return xc.Slice(slice)
}

func (xc *XCtx) SliceSort(slice interface{}) *XCtx {
	xc.sort()
	if sort := xc.GetMsgData().sort; sort != "" { //排序
		slice = getSortSlice(slice, sort)
	}
	return xc.Slice(slice)
}

// 从查询参数里面解析响应表格表头的键值对,根据查询顺序排列表头顺序,不能使用默认解析到map中(map没有顺序)
// download=1&page=1&page=2&name=姓名&name=姓名222&sex=男
func (xc *XCtx) excelTitle() [][]string {
	rawQuery, err := url.QueryUnescape(xc.Request.URL.RawQuery)
	assert(err)
	s := strings.Split(rawQuery, "&")
	ss := make([][]string, 0)
	for i := 0; i < len(s); i++ {
		kv := strings.Split(s[i], "=")
		if len(kv) != 2 {
			continue
		}
		if kv[0] == query_key_page || kv[0] == query_key_item || kv[0] == query_key_download || kv[0] == query_key_sort {
			continue
		}
		ss = append(ss, kv)
	}
	return ss
}

// 从msgData中的List生成文件g
func (xc *XCtx) writeExcelFile() {
	et := xc.excelTitle()
	log.Println("excel 表头:", et)
	ld, ok := xc.GetMsgData().Data.(*ListData)
	if !ok {
		panic("通过分页数据生成excel文件异常,分页数据不是数ListData对象")
	}
	xf := xlsx.NewFile()
	sheet, _ := xf.AddSheet("1")
	bs, err := json.Marshal(ld.List)
	if err != nil {
		panic("data marshal panic errmsg:" + err.Error())
	}
	ms := []map[string]interface{}{}
	if err = json.Unmarshal(bs, &ms); err != nil {
		panic("data Unmarshal panic errmsg:" + err.Error() + ",data:" + string(bs))
	}
	titleRow := sheet.AddRow()
	for i := 0; i < len(et); i++ {
		titleRow.AddCell().SetValue(et[i][1])
	}
	for _, m := range ms {
		row := sheet.AddRow()
		for _, v := range et {
			row.AddCell().SetValue(m[v[0]])
		}
	}
	xc.Header("content-type", "blob") //POST请求需要客户端指定
	xc.Header("Content-Disposition", "attachment;Filename="+"1.xlsx")
	xc.Header("Content-Disposition", "attachment;Filename="+time.Now().Format("2006-01-02.xlsx"))
	xf.Write(xc.Writer)
	xc.Status(http.StatusOK)
}

func (xc *XCtx) Write() {
	if xc.Query("download") == "1" { //以excel文件返回数据
		xc.writeExcelFile()
		return
	}
	switch xc.dataKind {
	case data_kind_json:
		xc.Context.JSON(xc.statusCode, xc.data)
	case data_kind_string:
		xc.Context.String(xc.statusCode, xc.data.(string))
	case data_kind_msg_data:
		xc.Context.JSON(xc.statusCode, xc.msgData)
	default:
		panic(fmt.Sprintf("不支持的数据响应格式类型:%d", xc.dataKind))
	}
}

// 配置数据库连接接口
func (xc *XCtx) IniConScope() {
	xc.Context.Set("conscope", "TODO")
	log.Println("xCtx iniConScope TODO")
}

func (xc *XCtx) ConScope() interface{} {
	return xc.Context.MustGet("conscope")
}
