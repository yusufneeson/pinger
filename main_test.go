package main

import (
	"context"
	"testing"
	"time"
)

type fakeChecker struct {
	results map[string]Result
	delay   map[string]time.Duration
}

func (f fakeChecker) Check(parentContext context.Context, url string) Result {
	time.Sleep(f.delay[url])
	return f.results[url]
}

func TestWorker(t *testing.T) {
	url := []string{"http://slow", "http://fast", "http://timeout"}

	fake := fakeChecker{results: map[string]Result{
		"http://slow":    {URL: "http://slow", Code: 200},
		"http://fast":    {URL: "http://fast", Code: 200},
		"http://timeout": {URL: "http://timeout", Err: context.DeadlineExceeded},
	}, delay: map[string]time.Duration{
		"http://slow": 1 * time.Second,
	}}

	got := worker(context.Background(), fake, url, 2)
	if len(got) != len(url) {
		t.Fatalf("len(results) got %d, want %d", len(got), len(url))
	}

	for i, u := range url {
		if got[i].URL != u {
			t.Fatalf("results[%d] = %s, want %s", i, got[i].URL, u)
		}
	}

	if got[0].OK() != got[1].OK() {
		t.Fatalf("http://slow and http://fast must be OK")
	}

	if got[2].OK() {
		t.Fatalf("http://timeout must be error")
	}
}
