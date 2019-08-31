package xgin

import "github.com/gin-gonic/gin"

// 配置数据库连接接口
func IniConScope(ctx *gin.Context) {
	Get(ctx).IniConScope()
}

// 自动写数据
func Write(ctx *gin.Context) {
	Get(ctx).Write()
}
