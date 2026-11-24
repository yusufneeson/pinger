package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
)

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
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
	fmt.Println(url, len(url))
}
