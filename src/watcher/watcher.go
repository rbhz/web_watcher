package watcher

import (
	"sync"
	"time"
)

// Notifier interface
type Notifier interface {
	Notify(updated []URL)
}

// Watcher check if urls changed
type Watcher struct {
	urls   []URL
	period int
}

// Check all urls
func (w *Watcher) Check() (res []URL) {
	var wg sync.WaitGroup
	wg.Add(len(w.urls))
	for idx := range w.urls {
		go func(idx int) {
			defer wg.Done()
			if updated := w.urls[idx].Check(); updated {
				res = append(res, w.urls[idx])
			}
		}(idx)
	}
	wg.Wait()
	return res
}

// Start watcher as daemon
func (w *Watcher) Start(notifiers []Notifier) {
	for {
		updated := w.Check()
		for _, n := range notifiers {
			n.Notify(updated)
		}
		time.Sleep(time.Duration(w.period) * time.Second)
	}
}

// GetUrls return watchers urls slice
func (w Watcher) GetUrls() []URL {
	return w.urls
}

// GetWatcher returns watcher
func GetWatcher(urls []string, period int) Watcher {
	checker := Watcher{period: period}
	for _, url := range urls {
		checker.urls = append(checker.urls, URL{Link: url})
	}
	return checker
}
