package watcher

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Changed fields
const (
	StatusChange = iota
	HashChange
	ErrorChange
)

// URL struct
type URL struct {
	id         int
	Link       string    `json:"url"`
	LastChange time.Time `json:"last_change"`
	Status     int       `json:"status"`
	Err        string    `json:"error"`
	lastCheck  time.Time
	hash       []byte
}

func (u *URL) log(level func() *zerolog.Event) *zerolog.Event {
	return level().Str("url", u.Link)
}

// Update url
func (u *URL) Update() URLUpdate {
	u.log(log.Debug).Msg("Updating")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(u.Link)
	if err != nil {
		return u.change([]byte{}, 0, err)
	}
	defer resp.Body.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, resp.Body); err != nil {
		return u.change([]byte{}, 0, err)
	}
	var hashSum []byte = hash.Sum(nil)
	return u.change(hashSum, resp.StatusCode, nil)
}

func (u *URL) change(hash []byte, status int, err error) URLUpdate {
	now := time.Now()
	old := *u
	var changes []int
	if bytes.Compare(u.hash, hash) != 0 {
		changes = append(changes, HashChange)
		u.hash = hash
	}
	if err != nil && err.Error() != u.Err {
		changes = append(changes, ErrorChange)
		u.Err = err.Error()
		u.Status = 0
	} else if status != u.Status {
		changes = append(changes, ErrorChange)
		u.Status = status
		u.Err = ""
	}
	u.lastCheck = now
	res := URLUpdate{
		New:     *u,
		Old:     old,
		Changed: changes,
		Created: now,
	}
	if len(changes) != 0 {
		u.LastChange = u.lastCheck
	}
	return res
}

func (u *URL) save(db *sql.DB) (err error) {
	stmt, err := db.Prepare("INSERT OR REPLACE INTO urls VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		u.log(log.Error).Err(err).Msg("Failed to prepare save statement")
		return
	}
	defer stmt.Close()
	res, err := stmt.Exec(u.Link, u.LastChange, u.hash, u.Status, u.Err)
	if err != nil {
		u.log(log.Error).Err(err).Msg("Failed to execute save statement")
		return
	}
	_, err = res.RowsAffected()
	if err != nil {
		u.log(log.Error).Err(err).Msg("No rows updates")
		return
	}
	u.log(log.Debug).Msg("Updated in database")
	return
}

// Good return true if last request was successfull
func (u URL) Good() bool {
	return u.Err == "" && u.Status == http.StatusOK
}

func getURL(id int, link string, db *sql.DB) *URL {
	url := &URL{
		id:        id,
		Link:      link,
		lastCheck: time.Now()}
	err := db.QueryRow(
		"SELECT last_change, hash, status, error FROM urls WHERE link=?;", url.Link,
	).Scan(&url.LastChange, &url.hash, &url.Status, &url.Err)
	if err != nil {
		if err == sql.ErrNoRows {
			url.Update()
			err = url.save(db)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to save url to DB")
			}
		} else {
			log.Fatal().Err(err).Msg("Failed to get url info from DB")

		}
	}
	return url
}

// URLUpdate contains information about url changes
type URLUpdate struct {
	New     URL
	Old     URL
	Changed []int
	Created time.Time
}

// Error return error description
func (u URLUpdate) Error() *string {
	if u.New.Err != "" {
		return &u.New.Err
	} else if u.New.Status != http.StatusOK {
		errText := strconv.Itoa(u.New.Status) + " status"
		return &errText
	}
	return nil
}
