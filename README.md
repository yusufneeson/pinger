# PINGER

Build CLI tools to check urls healthy. The result shound looks like:

```
$ pinger -f urls.txt -workers 5 -timeout 3s
OK    200  123ms  https://go.dev
OK    200   88ms  https://github.com
FAIL  ---  3000ms https://localhost:9999   (timeout)
FAIL  500   45ms  https://httpstat.us/500

4 checked, 2 ok, 2 failed
```

- Without any external packages
- Using worker pool
- Timeout per request
- Result should sequiental
- Without `panic` all error should explicitly
- Sparate I/O -> using interface
- Using test
- Shound passed from `go vet` and `go test -race`
