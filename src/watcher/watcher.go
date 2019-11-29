package watcher

import (
	"bytes"
	"crypto/md5"
	"io"
	"net/http"
	"sync"
	"time"
)

type Url struct {
	Url        string
	Good       bool
	LastUpdate time.Time
	LastCheck  time.Time
	lastHash   []byte
}

func (u *Url) Check() bool {
	tr := &http.Transport{
		IdleConnTimeout: 30 * time.Second,
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(u.Url)
	if err != nil {
		u.update([]byte{}, false)
		return u.LastCheck == u.LastUpdate
	}
	defer resp.Body.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, resp.Body); err != nil {
		u.update([]byte{}, false)
		return u.LastCheck == u.LastUpdate
	}
	var hashSum []byte = hash.Sum(nil)
	u.update(hashSum, true)
	return u.LastCheck == u.LastUpdate
}
func (u *Url) update(hash []byte, good bool) {
	now := time.Now()
	if bytes.Compare(u.lastHash, hash) != 0 || u.Good != good {
		u.LastUpdate = now
	}
	u.lastHash = hash
	u.Good = good
	u.LastCheck = now
}

type Watcher struct {
	urls   []Url
	period int
}

func (w *Watcher) Check() []int {
	var res []int
	var wg sync.WaitGroup
	wg.Add(len(w.urls))
	for idx := range w.urls {
		go func(idx int) {
			defer wg.Done()
			if updated := w.urls[idx].Check(); updated {
				res = append(res, idx)
			}
		}(idx)
	}
	wg.Wait()
	return res
}

func (w *Watcher) Start(callback func([]int)) {
	for {
		callback(w.Check())
		time.Sleep(time.Duration(w.period) * time.Second)
	}
}

func (w Watcher) GetUrls() []Url {
	return w.urls
}
func GetWatcher(urls []string, period int) Watcher {
	checker := Watcher{period: period}
	for _, url := range urls {
		checker.urls = append(checker.urls, Url{Url: url})
	}
	return checker
}
