
# Froxy Proxy

`froxy-proxy` is a simple web proxy that supports `HTTP Tunneling`.

## Build

```
go build .
```

## Usage
```
Usage of froxy-proxy:
  -addr string
        address to serve on (default ":5555")
  -allow string
        CONNECT allowed ports (comma separated list) (default "443")
```
`addr` is the address on which the server will run, it listens on all interfaces on port `5555` by default.

`allow` is a comma separated list of the allowed ports while using `CONNECT`, it contains only `443` by default.
```
./froxy-proxy
```
```
2022/09/10 16:47:08 Serving on :5555
2022/09/10 16:47:08 CONNECT allowed ports: [443]



```
Now the proxy is ready to be used.

## Notes
* Care must be taken while specifing the allowed ports as the proxy can be used for attacks.
  > There are significant risks in establishing a tunnel to arbitrary servers, particularly when the destination is a well-known or reserved TCP port that is not intended for Web traffic.  For example, a CONNECT to a request-target of "example.com:25" would suggest that the proxy connect to the reserved port for SMTP traffic; if allowed, that could trick the proxy into relaying spam email.  Proxies that support CONNECT SHOULD restrict its use to a limited set of known ports or a configurable whitelist of safe request targets.

* Binding to ports below 1024 requires root priviliges.

## REF
* Hypertext Transfer Protocol (HTTP/1.1) - Semantics and Content (rfc 7231): https://www.rfc-editor.org/rfc/rfc7231