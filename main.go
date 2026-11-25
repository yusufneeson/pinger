package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
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

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func checkUrl(parentContext context.Context, client *http.Client, url string, timeout int) ResultUrl {

	ctx, cancel := context.WithTimeout(parentContext, time.Duration(timeout*int(time.Second)))
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

func main() {
	file := flag.String("f", "", "File with list of urls")
	workers := flag.Int("workers", 3, "Total workers, default is 3")
	timeout := flag.Int("timeout", 3, "Max timeout of http request, default is 3second")

	flag.Parse()

	if flag.NFlag() < 1 {
		fmt.Println("Use with -f file.txt")
		return
	}
	loc := path.Join(path.Dir(*file), path.Base(*file))

	f, err := os.Open(loc)
	check(err)
	defer f.Close()

	scanner := bufio.NewScanner(f)

	var url []string

	for scanner.Scan() {
		url = append(url, scanner.Text())
	}

	check(scanner.Err())

	jobs := make(chan string, len(url))
	result := make(chan ResultUrl)

	for _, targetUrl := range url {
		jobs <- targetUrl
	}
	close(jobs)

	var wg sync.WaitGroup

	for i := 1; i <= *workers; i++ {
		wg.Go(func() {
			for url := range jobs {
				ctx := context.Background()
				client := &http.Client{Timeout: time.Duration(*timeout) * time.Second}
				res := checkUrl(ctx, client, url, *timeout)
				result <- res
			}
		})
	}

	ts := time.Now()

	go func() {
		wg.Wait()
		close(result)
	}()

	for r := range result {
		fmt.Println(r)
	}

	elapsed := time.Since(ts)
	fmt.Println(elapsed)
}
