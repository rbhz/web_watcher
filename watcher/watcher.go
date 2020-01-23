package watcher

import (
	"database/sql"
	"log"
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
	dbPath string
	db     *sql.DB
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
	if w.db == nil {
		w.initDB()
	}
	for range time.Tick(time.Duration(w.period) * time.Second) {
		updated := w.Check()
		for _, url := range updated {
			go url.save(w.db)
		}
		for _, n := range notifiers {
			go n.Notify(updated)
		}
	}
}

func (w *Watcher) initDB() {
	db, err := sql.Open("sqlite3", w.dbPath)
	if err != nil {
		log.Fatalf("Failed to open db: %v", err)
	}
	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS urls (
			link VARCHAR(200) PRIMARY KEY,
			last_change DATE NOT NULL,
			hash BLOB NOT NULL,
			good BOOL NOT NULL
		);`,
	)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
	w.db = db
}

// GetUrls return watchers urls slice
func (w Watcher) GetUrls() []URL {
	return w.urls
}

// NewWatcher returns watcher
func NewWatcher(urls []string, period int, db string) Watcher {
	watcher := Watcher{period: period, dbPath: db}
	watcher.initDB()
	var wg sync.WaitGroup
	wg.Add(len(urls))
	for _, url := range urls {
		go func(url string) {
			defer wg.Done()
			watcher.urls = append(watcher.urls, getURL(url, watcher.db))
		}(url)
	}
	wg.Wait()
	return watcher
}
