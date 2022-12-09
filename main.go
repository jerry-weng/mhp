package main

import (
	"flag"
)

type Opt struct {
	Config string
	Listen int
	Prefix string
	Root string
	Upstream string
	Version bool
}

var opt Opt

func init() {
	flag.StringVar(&opt.Config, "c", "", "config file")
	flag.IntVar(&opt.Listen, "l", 9000, "server listening port")
	flag.StringVar(&opt.Prefix, "prefix", "/", "prefix of request path, example: /api/")
	flag.StringVar(&opt.Root, "root", "", "root path of static files, example: dist")
	flag.StringVar(&opt.Upstream, "upstream", "", "upstream server, example: http://localhost:3000")
	flag.BoolVar(&opt.Version, "version", false, "show current version")
}

const VERSION = "0.2"

func main() {
	flag.Parse()

	if opt.Version {
		println(VERSION)
		return
	}
	if opt.Config != "" {
		config, err := ParseConfig(opt.Config)
		if err != nil {
			panic(err)
		}
		Proxy(config)
	} else {
		var config ProxyConfig
		var proxyItem ProxyItem
		config.Listen = opt.Listen
		proxyItem.Prefix = opt.Prefix
		if opt.Root != "" {
			proxyItem.Root = opt.Root
		} else if opt.Upstream != "" {
			proxyItem.Upstream = opt.Upstream
		}
		config.Proxy = []ProxyItem{proxyItem}
		Proxy(config)
	}
}
