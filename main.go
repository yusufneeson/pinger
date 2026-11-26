package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Result struct {
	url    string
	status string
	code   int
	time   time.Duration
	error  error
}

type Job struct {
	id  int
	url string
}

// -- I/O --
func readUrl(path string) ([]string, error) {
	var url []string
	f, err := os.Open(path)
	if err != nil {
		return url, err
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		url = append(url, scanner.Text())
	}

	return url, scanner.Err()
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "pinger:", err)
		os.Exit(1)
	}
}

func run() error {
	file := flag.String("f", "", "File with list of urls")
	workers := flag.Int("workers", 3, "Total workers, default is 3")
	timeout := flag.Duration("timeout", 3*time.Second, "Max timeout of http request, default is 3s")
	flag.Parse()

	if *file == "" {
		flag.Usage()
		return errors.New("-f file not found!")
	}

	if *workers < 1 {
		flag.Usage()
		return errors.New("-worker minimal 1")
	}

	url, e := readUrl(*file)
	if e != nil {
		return e
	}

	// -- time calculation: start --
	ts := time.Now()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	results := worker(ctx, url, *workers, *timeout)

	for _, r := range results {
		fmt.Println(r)
	}

	// -- time calculation: end --
	elapsed := time.Since(ts)
	fmt.Println(elapsed)

	return nil
}

func worker(ctx context.Context, url []string, workers int, timeout time.Duration) []Result {
	jobs := make(chan Job)
	results := make([]Result, len(url))

	client := &http.Client{}

	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for j := range jobs {
				results[j.id] = checkUrl(ctx, client, j.url, timeout)
			}
		})
	}

	for i, u := range url {
		jobs <- Job{i, u}
	}

	close(jobs)
	wg.Wait()

	return results
}

func checkUrl(parentContext context.Context, client *http.Client, url string, timeout time.Duration) Result {
	ctx, cancel := context.WithTimeout(parentContext, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return Result{url: url, status: "FAIL", error: err}
	}

	start := time.Now()
	resp, e := client.Do(req)
	elapsed := time.Since(start)
	if e != nil {
		return Result{url: url, status: "FAIL", time: elapsed, error: e}
	}

	defer resp.Body.Close()

	return Result{url: url, code: resp.StatusCode, status: http.StatusText(resp.StatusCode), time: elapsed}
}
