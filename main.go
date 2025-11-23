package main

import (
	"bufio"
	"context"
	"embed"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/energye/systray"
	"github.com/sirupsen/logrus"
	writer "github.com/sirupsen/logrus/hooks/writer"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

var app *App

func runWindow() {
	defer recover()

	err := wails.Run(&options.App{
		Title:  "Sync Folder",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown: func(ctx context.Context) {
			app = nil
		},
		// Menu:             app.applicationMenu(),
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func openWindow() {
	app = NewApp()
	runWindow()
}

func closeWindow() {
	defer recover()
	if app == nil || app.ctx == nil {
		app = nil
		return
	}
	runtime.Quit(app.ctx)
	app = nil
}

var logs []string = make([]string, 0)

type LogWriter struct {
	io.Writer
}

var lastCheckLogTime int64 = 0

const DAY_IN_SECONDS = 24 * 60 * 60

func (w *LogWriter) Write(p []byte) (n int, err error) {
	if windowStdin != nil {
		windowStdin.Write([]byte("Log:" + string(p) + "\n"))
	}

	// logs = append(logs, string(p))
	// if len(logs) > 100 {
	// 	logs = logs[1:]
	// }

	now := time.Now().Unix()
	if now-lastCheckLogTime > DAY_IN_SECONDS*3 {
		_, err = os.Stat(getLogPath() + ".old")
		if err == nil {
			os.Remove(getLogPath() + ".old")
		}
		os.Rename(getLogPath(), getLogPath()+".old")
		lastCheckLogTime = now
	}

	return len(p), nil
}
func readLog() {
	reader := bufio.NewReader(os.Stdin)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// 标准输出已关闭，退出循环
				break
			}
			// 处理其他错误
			logrus.Println("读取标准输出时出错:", err)
			return
		}
		// 处理读取到的每一行数据
		line = strings.TrimSpace(line)
		logrus.Println("读取到的行:", line)
		if strings.Index(line, "Log:") == 0 {
			logs = append(logs, line[4:])
			if len(logs) > 100 {
				logs = logs[1:]
			}
		}
	}

}
func runTray() {

	logFile, _ := os.OpenFile(getLogPath(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
	logrus.AddHook(&writer.Hook{
		Writer:    logFile,
		LogLevels: []logrus.Level{logrus.InfoLevel},
	})

	logrus.AddHook(&writer.Hook{
		Writer:    &LogWriter{},
		LogLevels: []logrus.Level{logrus.InfoLevel},
	})
	go syncConfigFolders(readConfig())

	// 每天执行一次全量同步
	ticker := time.NewTicker(time.Hour * 24)
	go func() {
		for range ticker.C {
			errors := ""
			for _, fp := range config.FolderPairs {
				err := diffAndSync(fp.Src, fp.Dst)
				if err != nil {
					errors += err.Error() + "\n"
				}
			}

			if errors != "" {
				logrus.Errorln("daily sync error:" + errors)
			}
		}
	}()
	systray.Run(onReady, onExit)
	ticker.Stop()
}

func main() {
	// go 获取 action 参数
	action := flag.String("action", "", "action")
	flag.Parse()
	switch *action {
	case "open-window":
		go readLog()
		fmt.Println("init")
		openWindow()
		return
	default:
		runTray()
		return
	}

}
