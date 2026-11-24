package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	args := os.Args[1:]
	file := args[0]
	loc := path.Join(path.Dir(file), path.Base(file))

	f, err := os.Open(loc)
	check(err)
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fmt.Printf("Line : %s\n", scanner.Text())
		fmt.Println(scanner.Text())
	}

	check(scanner.Err())

	fmt.Println("Hello", loc)
}
