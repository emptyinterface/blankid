
#blankid

a simple tool to replace unused function params with the blank id "_".

### Usage:

```
Usage: blankid [-w] [-recv] [-ret] <file[s]>
  -recv
      include receivers in scan
  -ret
      include returns in scan
  -w
      overwrite file with changes
```

###Before:

```go
http.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
})
```

###After:

```go
http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
})
```

License: MIT 2016