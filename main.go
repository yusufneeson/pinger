package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

type ResultUrl struct {
	url    string
	status string
	code   int
	time   int64
	error  string
}

type Job struct {
	id  int
	url string
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "pinger:", err)
		os.Exit(1)
	}
}

func checkUrl(parentContext context.Context, client *http.Client, url string, timeout time.Duration) ResultUrl {

	ctx, cancel := context.WithTimeout(parentContext, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()
	resp, e := client.Do(req)
	elapsed := time.Since(start)
	if e != nil {
		return ResultUrl{url: url, status: "FAIL", code: 0, time: elapsed.Milliseconds(), error: e.Error()}
	}

	defer resp.Body.Close()

	return ResultUrl{url: url, code: resp.StatusCode, status: http.StatusText(resp.StatusCode), time: elapsed.Milliseconds()}
}

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

	jobs := make(chan Job, len(url))
	results := make([]ResultUrl, len(url))

	for i, u := range url {
		jobs <- Job{id: i, url: u}
	}
	close(jobs)

	var wg sync.WaitGroup

	client := &http.Client{}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	for i := 1; i <= *workers; i++ {
		wg.Go(func() {
			for j := range jobs {
				res := checkUrl(ctx, client, j.url, *timeout)
				results[j.id] = res
			}
		})
	}

	ts := time.Now()
	wg.Wait()

	for _, r := range results {
		fmt.Println(r)
	}

	elapsed := time.Since(ts)
	fmt.Println(elapsed)

	return nil
}
