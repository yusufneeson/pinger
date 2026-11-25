package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

type ResultUrl struct {
	url    string
	status string
	code   int8
	time   int
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func checkUrl(url string) string {
	resp, e := http.Head(url)
	if e != nil {
		// log.Fatal(e)
		return "FAIL ___   " + url + " (" + e.Error() + ") "
	}

	defer resp.Body.Close()

	sts := strings.Split(resp.Status, " ")

	return sts[1] + "   " + sts[0] + "   " + url
}

func main() {
	file := flag.String("f", "", "File with list of urls")
	workers := flag.Int("workers", 3, "Total workers, default is 3")
	fmt.Printf("with %d workers\n", *workers)
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
	result := make(chan string)

	for _, targetUrl := range url {
		jobs <- targetUrl
	}
	close(jobs)

	var wg sync.WaitGroup

	for i := 1; i <= *workers; i++ {
		wg.Go(func() {
			for url := range jobs {
				res := checkUrl(url)
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
