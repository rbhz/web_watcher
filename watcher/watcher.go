package watcher

import (
	"database/sql"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// Notifier interface
type Notifier interface {
	Notify(URLUpdate)
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

	numWorkers := (len(w.urls) / 2) + 1
	log.Info().Int("count", numWorkers).Msg("Starting watcher workers")
	for i := 1; i <= numWorkers; i++ {
		go w.worker(forCheck, updates)
	}

	for {
		select {
		case <-time.Tick(100 * time.Millisecond):
			for _, url := range w.urls {
				var period time.Duration
				if url.Good() {
					period = w.period * time.Second
				} else {
					period = w.errorPeriod * time.Second
				}
				if url.lastCheck.Add(period).Before(time.Now()) {
					if _, ok := checking[url.id]; ok {
						log.Warn().Str("url", url.Link).Msg("Skipped because still in checking stage")
					}
					checking[url.id] = true
					log.Debug().Str("url", url.Link).Msg("Found url to check")
					forCheck <- url
				}
			}
		case update := <-updates:
			log.Debug().Str("url", update.New.Link).Msg("Checked")
			update.New.save(w.db)
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
	log.Debug().Msg("Worker started")
	for url := range in {
		log.Debug().Str("url", url.Link).Msg("Got url to check")
		out <- url.Update()
	}
}

func (w *Watcher) initDB() {
	db, err := sql.Open("sqlite3", w.dbPath)
	log.Info().Msg("Initializing Database")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open db")
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
		log.Fatal().Err(err).Msg("Failed to create table")
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
		go func(idx int, url string) {
			defer wg.Done()
			watcher.urls = append(watcher.urls, getURL(idx, url, watcher.db))
		}(idx, url)
	}
	wg.Wait()
	return watcher
}
