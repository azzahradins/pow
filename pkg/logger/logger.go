package logger

import (
	"io"
	"log"
	"os"
	"sync"
	"time"
)

var (
	instance *log.Logger
	logFile  *os.File
	once     sync.Once
)

func New(debug bool, destination string) error {
	var initErr error

	once.Do(func() {
		var out io.Writer

		if debug {
			out = os.Stdout
		} else {
			today := time.Now().Format("20060102")
			if destination == "" {
				destination = "./logs"
			}
			if _, err := os.Stat(destination); os.IsNotExist(err) {
				if err := os.MkdirAll(destination, 0755); err != nil {
					initErr = err
					return
				}
			}
			logFile, err := os.OpenFile(destination+"/"+today+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				initErr = err
				return
			}
			out = logFile
		}

		instance = log.New(out, "[POW] ", log.Ltime)
	})

	return initErr
}

func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

func Println(v ...any) {
	if instance != nil {
		instance.Println(v...)
	}
}

func Printf(format string, v ...any) {
	if instance != nil {
		instance.Printf(format, v...)
	}
}

func Fatalf(format string, v ...any) {
	if instance != nil {
		instance.Fatalf(format, v...)
	}
}
