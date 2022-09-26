package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

type FroxyProxy struct {
	address      string
	allowedPorts []string
	credentials  string
}

func NewFroxyProxy(address string, allowedPorts []string, cred string) *FroxyProxy {
	return &FroxyProxy{address: address, allowedPorts: allowedPorts, credentials: cred}
}

const (
	green = "\033[92m"
	// blue   = "\033[94m"
	// red    = "\033[0;31m"
	orange = "\033[38;5;214m"
	end    = "\033[0m"
)

func (s *FroxyProxy) Authenticate(r *http.Request) bool {
	if s.credentials == "" {
		return true // no auth
	}
	auth := r.Header.Get("Proxy-Authorization")
	if auth == "" {
		return false // Proxy-Authorization is not set
	}
	params := strings.Split(auth, " ")
	if params[0] != "Basic" {
		return false // Proxy-Authorization scheme is not "Basic"
	}
	if providedCred, err := base64.StdEncoding.DecodeString(params[1]); err != nil || string(providedCred) != s.credentials {
		return false // invalid credentials
	}
	return true // valid credentials
}

func (s *FroxyProxy) ListenAndServe() error {
	log.Println("Serving on", s.address)
	log.Println("CONNECT allowed ports:", s.allowedPorts)

	return http.ListenAndServe(s.address, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !s.Authenticate(r) {
			w.Header().Add("Proxy-Authenticate", "Basic")
			http.Error(w, "Proxy Authentication Required", http.StatusProxyAuthRequired)
			io.Copy(io.Discard, r.Body)
			r.Body.Close()
			return
		}

		switch r.Method {
		default:
			log.Printf("%s[%s]%s: %s - from %s\n", orange, r.Method, end, r.Host, r.RemoteAddr)
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
			w.WriteHeader(resp.StatusCode)

			// copy body
			_, err = io.Copy(w, resp.Body)
			if err != nil {
				log.Println(err)
			}

		case http.MethodConnect:
			log.Printf("%s[%s]%s: %s - from %s\n", green, r.Method, end, r.Host, r.RemoteAddr)
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
						return
					}
					defer clientConn.Close()
					// w.WriteHeader(http.StatusOK) puts "Transfer-Encoding: chunked" header
					// and this behaviour can't be avoided
					// for this reason I'm writing the status code response directly to the socket in this way:
					// clientConn.Write([]byte("HTTP/1.1 200 OK\r\n"))
					clientConn.Write([]byte("HTTP/1.1 200 Connection established\r\n"))
					clientConn.Write([]byte(fmt.Sprintf("Date: %s\r\n\r\n", time.Now().Format(http.TimeFormat))))

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

			http.Error(w, fmt.Sprintf("(%s) port is not allowed", portStr), http.StatusForbidden)
		}
	}))
}
