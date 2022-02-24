package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

func generateProxy(conf ProxyItem) http.Handler {
	if conf.Root != "" {
		// serve as static file server
		log.Printf("host static dir \"%s\" with preifx \"%s\"", conf.Root, conf.Prefix)
		f := http.FileServer(http.Dir(conf.Root))
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
