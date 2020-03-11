package watcher

import (
	"database/sql"
	"log"
	"sync"
	"time"
)

// Notifier interface
type Notifier interface {
	Notify(updated URLUpdate)
}

// Watcher check if urls changed
type Watcher struct {
	urls        []*URL
	period      time.Duration
	errorPeriod time.Duration
	dbPath      string
	db          *sql.DB
}

// Start watcher as daemon
func (w *Watcher) Start(notifiers []Notifier) {
	checking := make(map[int]bool)
	forCheck := make(chan *URL)
	updates := make(chan URLUpdate)

	for i := 0; i <= len(w.urls)/2; i++ {
		// Start workers
		go w.worker(forCheck, updates)
	}
	for {
		select {
		case <-time.Tick(100 * time.Millisecond):
			now := time.Now()
			for idx, url := range w.urls {
				var period time.Duration
				if url.Good() {
					period = w.period * time.Second
				} else {
					period = w.errorPeriod * time.Second
				}
				if _, ok := checking[idx]; !ok &&
					url.lastCheck.Add(period).Before(now) {
					checking[idx] = true
					forCheck <- url
				}
			}
		case update := <-updates:
			delete(checking, update.Old.id)
			if len(update.Changed) > 0 {
				for _, n := range notifiers {
					go n.Notify(update)
				}
			}
		}
	}
}

func (w *Watcher) worker(in <-chan *URL, out chan<- URLUpdate) {
	for url := range in {
		out <- url.Update()
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
			hash BLOB,
			status INT NOT NULL,
			error VARCHAR(500) NOT NULL
		);`,
	)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
	w.db = db
}

// GetUrls return watchers urls slice
func (w Watcher) GetUrls() []*URL {
	return w.urls
}

// NewWatcher returns watcher
func NewWatcher(urls []string, cfg Config) Watcher {
	watcher := Watcher{
		period:      cfg.Period,
		errorPeriod: cfg.ErrorPeriod,
		dbPath:      cfg.DBPath}
	watcher.initDB()
	var wg sync.WaitGroup
	wg.Add(len(urls))
	for idx, url := range urls {
		go func(url string) {
			defer wg.Done()
			watcher.urls = append(watcher.urls, getURL(idx, url, watcher.db))
		}(url)
	}
	wg.Wait()
	return watcher
}
