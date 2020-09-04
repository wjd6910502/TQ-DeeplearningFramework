package util

import (
    "log"
	  "fmt"
    "os"
    "io"
 )

var logger *log.Logger = nil

func GetLogger() *log.Logger {
    return logger
}

func Error(message interface{}) {
    logger.Println("[ERROR]" + fmt.Sprintf("%v",message))
}

func Debug(message interface{}) {
    logger.Println("[DEBUG]" + fmt.Sprintf("%v",message))
}

func Info(message  interface{}) {
	logger.Println("[INFO]" + fmt.Sprintf("%v",message))
}

func Infof(format string, msg ...interface{}) {
	logger.Println("[INFO]" + fmt.Sprintf(format,msg...))
}

func Initialize() {
    f,err  := os.Create("./log/debug.log")
    if err != nil {
        fmt.Printf("Failed to create log file: %v\n", err)
        logger = log.New(os.Stderr, "[Debug]", log.LstdFlags)
    } else {
        // 设置日志输出到文件,定义多个写入器
        writers := []io.Writer{ f, os.Stdout}
        fileAndStdoutWriter := io.MultiWriter(writers...)
        // 创建新的log对象
        logger = log.New(fileAndStdoutWriter, "[Debug]", log.Ldate|log.Ltime|log.Lshortfile)
    }
}
