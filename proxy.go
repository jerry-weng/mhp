package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

type (
	// FallbackResponseWriter wraps an http.Requesthandler and surpresses
	// a 404 status code. In such case a given local file will be served.
	FallbackResponseWriter struct {
		WrappedResponseWriter http.ResponseWriter
		FileNotFound          bool
	}
)

// Header returns the header of the wrapped response writer
func (frw *FallbackResponseWriter) Header() http.Header {
	return frw.WrappedResponseWriter.Header()
}

// Write sends bytes to wrapped response writer, in case of FileNotFound
// It surpresses further writes (concealing the fact though)
func (frw *FallbackResponseWriter) Write(b []byte) (int, error) {
	if frw.FileNotFound {
		return len(b), nil
	}
	return frw.WrappedResponseWriter.Write(b)
}

// WriteHeader sends statusCode to wrapped response writer
func (frw *FallbackResponseWriter) WriteHeader(statusCode int) {
	log.Printf("INFO: WriteHeader called with code %d\n", statusCode)
	if statusCode == http.StatusNotFound {
		// log.Printf("INFO: Setting FileNotFound flag\n")
		frw.FileNotFound = true
		return
	}
	frw.WrappedResponseWriter.WriteHeader(statusCode)
}

// AddFallbackHandler wraps the handler func in another handler func covering authentication
func AddFallbackHandler(handler http.HandlerFunc, root string, filename string) http.HandlerFunc {
	log.Printf("INFO: Creating fallback handler")
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("INFO: Wrapping response writer in fallback response writer")
		frw := FallbackResponseWriter{
			WrappedResponseWriter: w,
			FileNotFound:          false,
		}
		handler(&frw, r)
		if frw.FileNotFound {
			log.Printf("INFO: Serving fallback to %s", filename)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			http.ServeFile(w, r, filepath.Join(root, filename))
		}
	}
}

func generateProxy(conf ProxyItem) http.Handler {
	if conf.Root != "" {
		// serve as static file server
		log.Printf("host static dir \"%s\" with preifx \"%s\"", conf.Root, conf.Prefix)
		// f := http.FileServer(http.Dir(conf.Root))

		// ref: https://stackoverflow.com/a/54466375
		f := AddFallbackHandler(http.FileServer(http.Dir(conf.Root)).ServeHTTP, conf.Root, conf.Fallback)

		return f
	}
	if conf.Upstream != "" {
		// proxy to upstream server
		log.Printf("proxy preifx \"%s\" to \"%s\"", conf.Prefix, conf.Upstream)
		origin, _ := url.Parse(conf.Upstream)
		proxy := &httputil.ReverseProxy{Director: func(req *http.Request) {
			originHost := origin.Host
			req.Header.Add("X-Forwarded-Host", req.Host)
			req.Header.Add("X-Origin-Host", originHost)
			req.Host = originHost
			req.URL.Host = originHost
			req.URL.Scheme = origin.Scheme
			destPath := strings.TrimPrefix(req.URL.Path, conf.Prefix)
			req.URL.Path = origin.Path + destPath
		}, Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).Dial,
		}}
		return proxy
	}
	panic(errors.New("invalid proxy conf"))
}

func Proxy(conf ProxyConfig) {
	port := conf.Listen

	for _, conf := range conf.Proxy {
		proxy := generateProxy(conf)
		http.HandleFunc(conf.Prefix, func(w http.ResponseWriter, r *http.Request) {
			proxy.ServeHTTP(w, r)
		})
	}

	log.Printf("Starting Listening on port %d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
