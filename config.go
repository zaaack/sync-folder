package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

type Config struct {
	FolderPairs []FolderPair `json:"folder_pairs"`
}

type FolderPair struct {
	Src string `json:"src"`
	Dst string `json:"dst"`
}

func getConfigPath() string {
	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s/config.json", path.Dir(exePath))
}
func getLogPath() string {
	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s/data.log", path.Dir(exePath))
}

func readConfig() Config {
	config := Config{FolderPairs: make([]FolderPair, 0)}

	configBytes, err := os.ReadFile(getConfigPath())
	if err != nil {
		fmt.Printf("read config error: %s\n", err.Error())
	}

	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		fmt.Printf("parse config json error: %s\n", err.Error())
	}
	return config
}

func writeConfig(config Config) error {
	configPath := getConfigPath()
	configBytes, err := json.Marshal(config)
	if err != nil {
		fmt.Printf("stringify config json error: %s\n", err.Error())
		return err
	}
	err = os.WriteFile(configPath, configBytes, 0644)
	if err != nil {
		fmt.Printf("write config json error: %s\n", err.Error())
		return err
	}
	return nil
}

var watchers []*fsnotify.Watcher = make([]*fsnotify.Watcher, 0)

func syncConfigFolders(config Config) {
	for _, w := range watchers {
		w.Close()
	}
	errors := ""

	for _, fp := range config.FolderPairs {
		watcher, err := syncFolder(fp.Src, fp.Dst)
		if err != nil {
			errors += fmt.Sprintf("sync folder error: %s\n", err.Error())
		} else {
			watchers = append(watchers, watcher)
		}
	}
	if errors != "" {
		logrus.Error("syncConfigFolders:" + errors)
	}
}
