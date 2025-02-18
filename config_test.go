package main

import (
	"testing"
	"time"
)

func TestMain(t *testing.T) {

	// 每天执行一次全量同步
	ticker := time.NewTicker(time.Second * 5)
	go func() {
		for range ticker.C {
			syncConfigFolders(config)
		}
	}()

	// 模拟运行一段时间
	time.Sleep(time.Second * 20)
	ticker.Stop()
}
