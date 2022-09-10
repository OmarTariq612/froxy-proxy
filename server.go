package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

type FroxyProxy struct {
	address      string
	allowedPorts []string
}

func NewFroxyProxy(address string, allowedPorts []string) *FroxyProxy {
	return &FroxyProxy{address: address, allowedPorts: allowedPorts}
}

func (s *FroxyProxy) ListenAndServe() error {
	log.Println("Serving on", s.address)
	log.Println("CONNECT allowed ports:", s.allowedPorts)

	return http.ListenAndServe(s.address, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%v]: %v - from %v\n", r.Method, r.Host, r.RemoteAddr)
		switch r.Method {
		default:
			resp, err := http.DefaultTransport.RoundTrip(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
			defer resp.Body.Close()

			// copy headers
			for key, values := range resp.Header {
				for _, value := range values {
					w.Header().Add(key, value)
				}
			}
			// copy body
			_, err = io.Copy(w, resp.Body)
			if err != nil {
				log.Println(err)
			}

		case http.MethodConnect:
			io.Copy(io.Discard, r.Body)
			_, portStr, _ := net.SplitHostPort(r.Host)
			for _, port := range s.allowedPorts {
				if portStr == port {
					serverConn, err := net.DialTimeout("tcp", r.Host, 7*time.Second)
					if err != nil {
						http.Error(w, err.Error(), http.StatusServiceUnavailable)
						return
					}
					defer serverConn.Close()
					// w.WriteHeader(http.StatusOK)

					hj, ok := w.(http.Hijacker)
					if !ok {
						http.Error(w, err.Error(), http.StatusServiceUnavailable)
						return
					}
					clientConn, _, err := hj.Hijack()
					if err != nil {
						http.Error(w, err.Error(), http.StatusServiceUnavailable)
					}
					defer clientConn.Close()
					// w.WriteHeader(http.StatusOK) puts "Transfer-Encoding: chunked" header
					// and this behaviour can't be avoided
					// for this reason I'm writing the status code response directly to the socket in this way:
					clientConn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

					errc := make(chan error, 2)
					go func() {
						_, err := io.Copy(serverConn, clientConn)
						if err != nil {
							err = fmt.Errorf("could not copy from client to server, %v", err)
						}
						errc <- err
					}()
					go func() {
						_, err := io.Copy(clientConn, serverConn)
						if err != nil {
							err = fmt.Errorf("could not copy from server to client, %v", err)
						}
						errc <- err
					}()
					err = <-errc
					if err != nil {
						log.Println(err)
					}
					return
				}
			}
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
}
