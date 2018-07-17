# httpproxy

- http client request proxy tools

## build
```
go build .
```

## usage

```
Usage of httpproxy.exe:
  -h    this help.
  -in string
        listen addr by http proxy.
  -out string
        http proxy redirect to addr.
  -time int
        http proxy run time.
```

## example

- forward proxy

```
httpproxy.exe -in 127.0.0.1:8080 -out "10.10.0.1:8080;10.10.0.2:8080"
```

- Reverse Proxy

```
httpproxy.exe -in :8080 -out "127.0.0.1:8080;127.0.0.2:8080"
```
