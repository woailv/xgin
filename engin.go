package xgin

import (
	"github.com/gin-gonic/gin"
)

type iniResourcer interface {
	IniResource(resName, oper, relativePath string) //初始化一条资源
	IniSuccess()                                    //资源初始化成功之后要做的操作
}

type xEngin struct {
	*gin.Engine
	iniResourcer iniResourcer
	resources    []string //已初始化的资源,重复则产生异常
}

func Engin(engin *gin.Engine) func(*xEngin) {
	return func(xe *xEngin) {
		xe.Engine = engin
	}
}

func New(confFunc ...func(*xEngin)) *xEngin {
	xe := new(xEngin)
	for _, f := range confFunc {
		f(xe)
	}
	if xe.Engine == nil {
		xe.Engine = gin.Default()
	}
	return xe
}

func (xe *xEngin) IniResourcer(iniResourcer iniResourcer) *xEngin {
	xe.iniResourcer = iniResourcer
	return xe
}

func (xe *xEngin) Run(addr ...string) error {
	if xe.iniResourcer != nil {
		xe.iniResourcer.IniSuccess()
	}
	return xe.Engine.Run(addr...)
}

func (xe *xEngin) RunTLS(addr, certFile, keyFile string) error {
	if xe.iniResourcer != nil {
		xe.iniResourcer.IniSuccess()
	}
	return xe.Engine.RunTLS(addr, certFile, keyFile)
}

type Resource struct {
	ResName      string
	Oper         string
	RelativePath string
}

func (xe *xEngin) Handle(resName, oper, method, relativePath string, handlers ...gin.HandlerFunc) {
	xe.iniRouter(method, relativePath, handlers...)
	xe.iniResource(resName, oper, relativePath)
}

func (xe *xEngin) iniRouter(method, relativePath string, handlers ...gin.HandlerFunc) {
	handlers = append([]gin.HandlerFunc{IniConScope}, append(handlers, Write)...)
	xe.Engine.Handle(method, relativePath, handlers...)
}

func (xe *xEngin) iniResource(resName, oper, relativePath string) {
	if xe.iniResourcer != nil {
		xe.iniResourcer.IniResource(resName, oper, relativePath)
		resource := resName + "." + oper
		for _, v := range xe.resources {
			if resource == v {
				panic("资源重复:" + resource)
			}
		}
		xe.resources = append(xe.resources, resource)
	}
}
