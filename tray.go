package main

import (
	"bufio"
	"embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/energye/systray"
)

//go:embed build/trayicon.ico
var trayicon embed.FS
var window *exec.Cmd
var windowStdout io.ReadCloser
var windowStdin io.WriteCloser

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
				fmt.Println("读取标准输出时出错:", err)
				return
			}
			line = strings.TrimSpace(line)
			// 处理读取到的每一行数据
			fmt.Println("读取到的行:", line+"---")
			if line == "save" {
				syncConfigFolders(readConfig())
			} else if line == "init" {
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
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")
	mQuit.Click(onExit)
}

func onExit() {
	// clean up here
	cancelWindowProcess()
	systray.Quit()
	os.Exit(0)
}
