package watcher

import (
	"bytes"
	"crypto/md5"
	"io"
	"net/http"
	"time"
)

// URL struct
type URL struct {
	Link       string    `json:"url"`
	Good       bool      `json:"good"`
	LastChange time.Time `json:"last_change"`
	LastCheck  time.Time `json:"-"`
	lastHash   []byte
}

// Check if url data changed
func (u *URL) Check() bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(u.Link)
	if err != nil {
		u.update([]byte{}, false)
		return u.LastCheck == u.LastChange
	}
	defer resp.Body.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, resp.Body); err != nil {
		u.update([]byte{}, false)
		return u.LastCheck == u.LastChange
	}
	var hashSum []byte = hash.Sum(nil)
	u.update(hashSum, true)
	return u.LastCheck == u.LastChange
}
func (u *URL) update(hash []byte, good bool) {
	now := time.Now()
	if bytes.Compare(u.lastHash, hash) != 0 || u.Good != good {
		u.LastChange = now
	}
	u.lastHash = hash
	u.Good = good
	u.LastCheck = now
}
