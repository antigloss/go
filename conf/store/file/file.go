package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/antigloss/go/conf/store"
)

// New 创建从指定文件中读取配置的 Store 对象。
func New(opts ...option) store.Store {
	a := &fileStore{}
	a.opts.apply(opts...)
	return a
}

type fileStore struct {
	opts options
}

// Load 加载配置
func (a *fileStore) Load() ([]store.ConfigContent, error) {
	paths, err := a.calculateFilePaths()
	if err != nil {
		return nil, err
	}

	contents := make([]store.ConfigContent, len(paths))
	for i, p := range paths {
		contents[i].Type, err = store.ConfigType(p)
		if err != nil {
			return nil, err
		}

		contents[i].Content, err = os.ReadFile(p)
		if err != nil {
			return nil, err
		}

		if a.opts.tData != nil {
			contents[i].Content, err = a.opts.tData.Replace(contents[i].Content)
			if err != nil {
				return nil, fmt.Errorf("%s: %s", err.Error(), p)
			}
		}
	}
	return contents, nil
}

// Watch 监听配置变化。暂时不支持该操作，直接返回 nil
func (a *fileStore) Watch(ch chan<- *store.ConfigChanges) error {
	return nil
}

// Unwatch 取消监听
func (a *fileStore) Unwatch() {
}

func (a *fileStore) calculateFilePaths() ([]string, error) {
	var paths []string

	for _, p := range a.opts.paths {
		f, err := os.Stat(p.Path)
		if err != nil {
			return nil, err
		}

		if f.IsDir() {
			ps, e := readDir(p.Path, p.Recursive)
			if e != nil {
				return nil, err
			}
			paths = append(paths, ps...)
		} else {
			paths = append(paths, p.Path)
		}
	}

	return paths, nil
}

func readDir(dir string, recursive bool) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".") {
			continue
		}

		if !file.IsDir() {
			paths = append(paths, filepath.Join(dir, file.Name()))
			continue
		}

		if !recursive {
			continue
		}

		ps, e := readDir(filepath.Join(dir, file.Name()), true)
		if e != nil {
			return nil, e
		}
		paths = append(paths, ps...)
	}
	return paths, nil
}
