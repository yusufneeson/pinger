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
)

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func get(url string) string {
	resp, e := http.Get(url)
	if e != nil {
		// log.Fatal(e)
		return "FAIL ___   " + url + " (" + e.Error() + ") "
	}

	defer resp.Body.Close()

	sts := strings.Split(resp.Status, " ")

	return sts[1] + "   " + sts[0] + "   " + url
}

func main() {
	file := flag.String("url", "", "File with list of urls")
	flag.Parse()

	if flag.NFlag() < 1 {
		fmt.Println("Use with -url file.txt")
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

	for _, u := range url {
		resp := get(u)
		fmt.Println(resp)
	}

}
