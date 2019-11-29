package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"watcher"
)

type Arguments struct {
	filePath string
	period   int
}

func getArguments() (args Arguments) {
	flag.Usage = func() {
		fmt.Printf("Usage: %s [OPTIONS] path_to_file \n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.IntVar(&args.period, "period", 10, "Update period")
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
	watcher := watcher.GetWatcher(readFile(args.filePath), args.period)
	watcher.Start(func(new []int) {
		urls := watcher.GetUrls()
		for _, idx := range new {
			fmt.Printf("%v updated\n", urls[idx].Url)
		}
	})
}
