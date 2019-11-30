package watcher

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"io"
	"log"
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
func (u *URL) save(db *sql.DB) (err error) {
	stmt, err := db.Prepare("INSERT OR REPLACE INTO urls VALUES(?, ?, ?, ?)")
	if err != nil {
		return
	}
	res, err := stmt.Exec(u.Link, u.LastChange, u.lastHash, u.Good)
	if err != nil {
		return
	}
	_, err = res.RowsAffected()
	if err != nil {
		return
	}
	return
}

func getURL(link string, db *sql.DB) (url URL) {
	url.Link = link
	err := db.QueryRow(
		"SELECT last_change, hash, good FROM urls WHERE link=?;", url.Link,
	).Scan(&url.LastChange, &url.lastHash, &url.Good)
	if err != nil {
		if err == sql.ErrNoRows {
			url.Check()
			err = url.save(db)
			if err != nil {
				log.Fatalf("Failed to save url to DB: %v", err)
			}
		} else {
			log.Fatalf("Failed to get url info from DB: %v", err)

		}
	}
	return
}
