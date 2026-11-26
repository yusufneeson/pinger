package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"text/tabwriter"
	"time"
)

type Result struct {
	URL      string
	Code     int
	Duration time.Duration
	Err      error
}

func (r Result) OK() bool { return r.Err == nil && r.Code < 400 }

func (r Result) Status() string {
	if r.Err != nil {
		return "FAIL"
	} else {
		return "OK"
	}
}

func (r Result) CodeText() string {
	if r.Code == 0 {
		return "---"
	}

	return strconv.Itoa(r.Code)
}

func (r Result) Note() string {
	var dnsErr *net.DNSError
	switch {
	case r.Err == nil:
		return ""
	case errors.Is(r.Err, context.DeadlineExceeded):
		return "(timeout)"
	case errors.As(r.Err, &dnsErr):
		return "(DNS error: " + dnsErr.Err + ")"
	default:
		return "(" + r.Err.Error() + ")"
	}
}

// -- Checker URL --

type Checker interface {
	Check(parentContext context.Context, url string) Result
}

type HTTPCheck struct {
	client  *http.Client
	timeout time.Duration
}

func (c HTTPCheck) Check(parentContext context.Context, url string) Result {
	ctx, cancel := context.WithTimeout(parentContext, c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return Result{URL: url, Err: err}
	}

	start := time.Now()
	resp, e := c.client.Do(req)
	elapsed := time.Since(start)
	if e != nil {
		return Result{URL: url, Duration: elapsed, Err: e}
	}

	defer resp.Body.Close()

	return Result{URL: url, Code: resp.StatusCode, Duration: elapsed}
}

// -- worker pool --
type Job struct {
	id  int
	url string
}

func worker(ctx context.Context, c Checker, url []string, workers int) []Result {
	jobs := make(chan Job)
	results := make([]Result, len(url))

	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for j := range jobs {
				results[j.id] = c.Check(ctx, j.url)
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

func report(w io.Writer, result []Result) {
	var failed int

	tw := tabwriter.NewWriter(w, 0, 8, 2, ' ', 0)
	for _, r := range result {
		fmt.Fprintf(tw, "%s\t%s\t%dms\t%s\t\t%s\n", r.Status(), r.CodeText(), r.Duration.Milliseconds(), r.URL, r.Note())
		if !r.OK() {
			failed++
		}
	}
	tw.Flush()

	fmt.Printf("\n%d checked, %d ok, %d failed\n", len(result), len(result)-failed, failed)
}

// -- main --
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

	url, err := readUrl(*file)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	check := HTTPCheck{client: &http.Client{}, timeout: *timeout}
	results := worker(ctx, check, url, *workers)
	report(os.Stdout, results)

	return nil
}
