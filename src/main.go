package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"notifiers"
	"os"
	"watcher"
	"web"

	_ "github.com/mattn/go-sqlite3"
)

type arguments struct {
	filePath   string
	period     int
	port       int
	staticPath string
	dbPath     string
}

func getArguments() (args arguments) {
	flag.Usage = func() {
		fmt.Printf("Usage: %s [OPTIONS] path_to_file \n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.IntVar(&args.period, "period", 10, "Update period")
	flag.IntVar(&args.port, "port", 8080, "Web server port")
	flag.StringVar(&args.staticPath, "static", "src/web/static", "Path to static folder")
	flag.StringVar(&args.dbPath, "sqlite", "./watcher.db", "Path to sqlite file")
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}
	args.filePath = flag.Arg(0)
	return
}

func readFile(path string) (lines []string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return
}

func main() {
	args := getArguments()
	watcherInstance := watcher.GetWatcher(
		readFile(args.filePath),
		args.period,
		args.dbPath,
	)

	// Start web interface
	webServer := web.GetServer(watcherInstance, args.staticPath, args.port)
	go webServer.Run()
	var ns []watcher.Notifier
	ns = append(ns, notifiers.WebNotifier{
		Server: webServer,
	})

	watcherInstance.Start(ns)
}
