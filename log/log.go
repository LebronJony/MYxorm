package log

/*
	为什么不直接使用原生的 log 库呢？log 标准库没有日志分级，不打印文件和行号，
	这就意味着我们很难快速知道是哪个地方发生了错误。

	这个简易的 log 库具备以下特性：

	1. 支持日志分级（Info、Error、Disabled 三级）。
	2. 不同层级日志显示时使用不同的颜色区分。
	3. 显示打印日志代码对应的文件名和行号。
*/

import (
	"io/ioutil"
	"log"
	"os"
	"sync"
)

// 创建 2 个日志实例分别用于打印 Info 和 Error 日志
var (
	// 颜色红色 log.Lshortfile 支持显示文件名和代码行号
	errorLog = log.New(os.Stdout, "\033[31m[erroer]\033[0m", log.LstdFlags|log.Lshortfile)
	// 颜色蓝色
	infoLog = log.New(os.Stdout, "\033[34m[info ]\033[0m ", log.LstdFlags|log.Lshortfile)
	// 指针数组
	loggers = []*log.Logger{errorLog, infoLog}
	mu      sync.Mutex
)

// 暴露 Error，Errorf，Info，Infof 4个方法
var (
	Error  = errorLog.Println
	Errorf = errorLog.Printf
	Info   = infoLog.Println
	Infof  = infoLog.Printf
)

// 支持设置日志的层级(InfoLevel, ErrorLevel, Disabled)
const (
	InfoLevel = iota
	ErrorLevel
	Disabled
)

func SetLevel(level int) {
	mu.Lock()
	defer mu.Unlock()

	for _, logger := range loggers {
		logger.SetOutput(os.Stdout)
	}

	// 如果该层级小于当前level，就不打印对应日志
	if ErrorLevel < level {
		errorLog.SetOutput(ioutil.Discard)
	}

	if InfoLevel < level {
		infoLog.SetOutput(ioutil.Discard)
	}
}
