package main

import (
	"bufio"
	"embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/energye/systray"
	"github.com/sirupsen/logrus"
)

//go:embed build/trayicon.ico
var trayicon embed.FS
var window *exec.Cmd
var windowStdout io.ReadCloser
var windowStdin io.WriteCloser
var config Config = readConfig()

func execWindowProcess() {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	window = exec.Command(exe, "-action", "open-window")
	windowStdout, err = window.StdoutPipe()
	if err != nil {
		panic(err)
	}
	windowStdin, err = window.StdinPipe()
	if err != nil {
		panic(err)
	}

	go func() {
		reader := bufio.NewReader(windowStdout)
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
			line = strings.TrimSpace(line)
			// 处理读取到的每一行数据
			fmt.Println("读取到的行:", line+"---")
			if line == "save" {
				config = readConfig()
				syncConfigFolders(config)
			} else if line == "init" {
				logs, _ := tailFile(getLogPath(), 100)
				for _, log := range logs {
					windowStdin.Write([]byte("Log:" + log + "\n"))
				}

			}
		}
	}()
	window.Run()
	window = nil
}

func cancelWindowProcess() {
	if window != nil {
		if window.Process != nil {
			window.Process.Kill()
		}
		windowStdin = nil
		windowStdout = nil
		window = nil
	}
}

func toggleWindow() {
	defer recover()
	if window == nil {
		go execWindowProcess()
	} else {
		cancelWindowProcess()
	}
}

func onReady() {
	icon, err := trayicon.ReadFile("build/trayicon.ico")
	if err != nil {
		panic(err)
	}

	systray.SetIcon(icon)
	systray.SetTitle("Sync folders")
	// systray.SetTooltip("")
	systray.SetOnClick(func(menu systray.IMenu) {
		toggleWindow()
	})
	systray.SetOnRClick(func(menu systray.IMenu) {
		menu.ShowMenu()
	})
	mShow := systray.AddMenuItem("Toggle", "Toggle the window")
	mShow.Click(func() {
		toggleWindow()
	})
	mRestart := systray.AddMenuItem("Restart", "Restart the app")
	mRestart.Click(func() {
		cancelWindowProcess()
		restart()
	})
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")
	mQuit.Click(onExit)

}

func onExit() {
	// clean up here
	cancelWindowProcess()
	systray.Quit()
	os.Exit(0)
}

func restart() {
	logrus.Info("重启程序...")

	// 获取当前可执行文件的路径
	exePath, err := os.Executable()
	if err != nil {
		logrus.Error("无法获取可执行文件路径:", err)
		return
	}

	// 使用 syscall.Exec 来替换当前进程
	err = syscall.Exec(exePath, os.Args, os.Environ())
	if err != nil {
		logrus.Error("重启失败:", err)
	}
}

// tailFile 读取文件的最后 n 行
func tailFile(filename string, n int) ([]string, error) {
	var lines []string
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
		lines = append(lines, scanner.Text())
		if lineCount > n {
			lines = lines[1:]
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
