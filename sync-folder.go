package main

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/hymkor/trash-go"
	"github.com/sirupsen/logrus"
)

func isFileOrSymbolicLink(info fs.FileInfo) bool {
	return info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0
}

func syncFile(srcPath, distPath string, info fs.FileInfo) error {
	if info == nil {
		_info, err := os.Stat(srcPath)
		if err != nil {
			return err
		}
		info = _info
	}

	if isFileOrSymbolicLink(info) {
		if distInfo, err := os.Stat(distPath); err == nil {
			if info.Size() != distInfo.Size() || info.ModTime() != distInfo.ModTime() {
				if err := copyFile(srcPath, distPath); err != nil {
					panic(err)
				}
				if err := os.Chtimes(distPath, info.ModTime(), info.ModTime()); err != nil {
					panic(err)
				}
				logrus.Infof("change %s", distPath)
			}
		} else {
			if err := copyFile(srcPath, distPath); err != nil {
				return err
			}
			if err := os.Chtimes(distPath, info.ModTime(), info.ModTime()); err != nil {
				return err
			}
			logrus.Infof("add %s", distPath)
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if _, err := os.Stat(dst); err == nil {
		err := os.RemoveAll(dst)
		if err != nil {
			panic(err)
		}
	}
	dstDir := filepath.Dir(dst)
	if _, err := os.Stat(dstDir); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return err
		}
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func getDistPathFromSrc(src, dist, srcPath string) string {
	rel, _ := filepath.Rel(src, srcPath)
	return filepath.Join(dist, rel)
}

func trashFile(distPath string) {
	logrus.Infof("try_trash %s", distPath)
	err := trash.Throw(distPath)
	if err != nil {
		logrus.Errorf("trash failed: %v, trying rm", err)
		err = os.RemoveAll(distPath)
		if err != nil {
			logrus.Panic("trash failed", err)
		}
		logrus.Infof("rm %s", distPath)
	} else {
		logrus.Infof("trash %s", distPath)
	}
}

func renameFile(src, dist, srcPath, oldDistPath string) {
	distPath := getDistPathFromSrc(src, dist, srcPath)
	if err := os.MkdirAll(filepath.Dir(distPath), 0755); err != nil {
		panic(err)
	}
	if err := os.Rename(oldDistPath, distPath); err != nil {
		panic(err)
	}
	logrus.Infof("rename %s to %s", oldDistPath, distPath)
}

func checkAndRename(src, dist, srcPath, oldDistPath string) bool {
	if _, err := os.Stat(oldDistPath); err == nil {
		distInfo, err := os.Stat(oldDistPath)
		if err != nil {
			panic(err)
		}
		srcInfo, err := os.Stat(srcPath)
		if err != nil {
			panic(err)
		}
		if (isFileOrSymbolicLink(srcInfo) && srcInfo.Size() == distInfo.Size()) ||
			(srcInfo.IsDir() && isFolderEqual(srcPath, oldDistPath)) {
			renameFile(src, dist, srcPath, oldDistPath)
			return true
		}
	}
	return false
}

func diffAndSync(src, dist string) error {
	start := time.Now()
	distFilesFromSrc := make(map[string]struct{})
	srcPathItems := []struct {
		path     string
		info     fs.FileInfo
		distPath string
	}{}
	distPathItems := []struct {
		path string
		info fs.FileInfo
	}{}

	filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		srcPathItems = append(srcPathItems, struct {
			path     string
			info     fs.FileInfo
			distPath string
		}{
			path:     path,
			info:     info,
			distPath: getDistPathFromSrc(src, dist, path),
		})
		return nil
	})

	filepath.Walk(dist, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		distPathItems = append(distPathItems, struct {
			path string
			info fs.FileInfo
		}{
			path: path,
			info: info,
		})
		return nil
	})

	srcPathItemsSet := make(map[string]struct{})
	for _, item := range srcPathItems {
		srcPathItemsSet[item.distPath] = struct{}{}
	}

	distPathItemsSet := make(map[string]struct{})
	for _, item := range distPathItems {
		distPathItemsSet[item.path] = struct{}{}
	}

	addedPathItems := []struct {
		path     string
		info     fs.FileInfo
		distPath string
	}{}
	for _, item := range srcPathItems {
		if _, ok := distPathItemsSet[item.distPath]; !ok {
			addedPathItems = append(addedPathItems, item)
		}
	}

	deletedPathItems := []struct {
		path string
		info fs.FileInfo
	}{}
	for _, item := range distPathItems {
		if _, ok := srcPathItemsSet[item.path]; !ok {
			deletedPathItems = append(deletedPathItems, item)
		}
	}

	for _, ap := range addedPathItems {
		if isFileOrSymbolicLink(ap.info) {
			for _, dp := range deletedPathItems {
				if ap.info.Size() > 10*1024*1024 && ap.info.Size() == dp.info.Size() {
					renameFile(src, dist, ap.path, dp.path)
					break
				}
			}
		}
	}

	for _, item := range srcPathItems {
		distPath := getDistPathFromSrc(src, dist, item.path)
		logrus.Infof("distPath: %s %s %s %s", src, dist, item.path, distPath)
		distFilesFromSrc[distPath] = struct{}{}
		if isFileOrSymbolicLink(item.info) {
			if err := syncFile(item.path, distPath, item.info); err != nil {
				return err
			}
		}
	}

	for _, item := range distPathItems {
		distPath := getDistPathFromSrc(dist, dist, item.path)
		if _, ok := distFilesFromSrc[distPath]; !ok {
			trashFile(distPath)
		}
	}

	logrus.Infof("diffAndSync %s took %v", src, time.Since(start))
	return nil
}

func isFolderEqual(src, dist string) bool {
	isSame := true
	err := filepath.Walk(src, func(p string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		logrus.Infof("walk:%s", p)
		if isFileOrSymbolicLink(info) {
			distPath := filepath.Join(dist, filepath.Base(p))
			distInfo, err := os.Stat(distPath)
			if err != nil {
				isSame = false
				return filepath.SkipDir
			}
			if info.Size() != distInfo.Size() {
				isSame = false
				return filepath.SkipDir
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return isSame
}

func syncFolder(src, dist string) (*fsnotify.Watcher, error) {
	if err := diffAndSync(src, dist); err != nil {
		return nil, err
	}

	// TODO: fork fsnotify 修改 enableRecurse
	// TODO: 重构重命名逻辑
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	renamedFile := ""
	go func() {
		defer recover()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				logrus.Infof("watchEvent %s %s", event.Op, event.Name)
				if event.Name == "" {
					continue
				}

				srcPath := event.Name
				distPath := getDistPathFromSrc(src, dist, srcPath)
				if event.Op.Has(fsnotify.Create) {
					srcInfo, err := os.Stat(srcPath)
					if err != nil {
						panic(err)
					}

					if checkAndRename(src, dist, srcPath, renamedFile) {
						continue
					}
					renamedFile = ""
					if isFileOrSymbolicLink(srcInfo) {
						if err := syncFile(srcPath, distPath, nil); err != nil {
							logrus.Errorf("syncFile failed: %v", err)
						}
					}
				} else if event.Op.Has(fsnotify.Remove) {
					trashFile(distPath)
				} else if event.Op.Has(fsnotify.Write) {
					if err := syncFile(srcPath, distPath, nil); err != nil {
						logrus.Errorf("syncFile failed: %v", err)
					}
				} else {
					if event.Op.Has(fsnotify.Rename) {
						if _, err := os.Stat(srcPath); os.IsNotExist(err) {
							renamedFile = distPath
							continue
						}
					}
					if err := syncFile(srcPath, distPath, nil); err != nil {
						logrus.Errorf("syncFile failed: %v", err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logrus.Errorf("watcher error: %v", err)
			}
		}
	}()

	err = watcher.Add(filepath.Join(src, "..."))
	if err != nil {
		return nil, err
	}
	return watcher, nil
}
