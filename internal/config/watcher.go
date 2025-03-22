package config

import (
	"context"
	"fmt"
	"github.com/lccxxo/bailuoli/internal/model"
	"log"
	"path/filepath"
	"sync"
	"time"
)
import "github.com/fsnotify/fsnotify"

type WatchCallback func(cfg *model.Config)

var (
	watcherLock sync.Mutex
	watchers    = make(map[string]*fsnotify.Watcher)
)

func Watch(ctx context.Context, path string, cb WatchCallback, debounce time.Duration) {
	watcherLock.Lock()
	defer watcherLock.Unlock()

	if _, ok := watchers[path]; ok {
		return
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		panic(fmt.Sprintf("Failed to create watcher: %v", err))
	}
	watchers[path] = w

	// 监听文件所在目录
	dir := filepath.Dir(path)
	if err := w.Add(dir); err != nil {
		panic(fmt.Sprintf("Failed to watch directory: %v", err))
	}

	go func() {
		defer w.Close()
		var lastUpdate time.Time
		timer := time.NewTimer(debounce)
		defer timer.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-w.Events:
				if !ok {
					return
				}
				// 仅处理写入事件
				if event.Op&fsnotify.Write == fsnotify.Write && event.Name == path {
					// 去抖动处理
					if time.Since(lastUpdate) < debounce {
						continue
					}
					lastUpdate = time.Now()

					// 加载新配置
					newCfg, err := Load(path)
					if err != nil {
						log.Printf("Config reload failed: %v", err)
						continue
					}

					// 触发回调
					cb(newCfg)
				}
			case err, ok := <-w.Errors:
				if !ok {
					return
				}
				log.Printf("Watcher error: %v", err)
			}
		}
	}()
}
