package main

import (
	"encoding/json"
	"io/ioutil"
)

type ProxyItem struct {
	Prefix   string `json:"prefix"`
	Root     string `json:"root,omitempty"`
	Fallback string `json:"fallback,omitempty"`
	Upstream string `json:"upstream,omitempty"`
}

type ProxyConfig struct {
	Version int         `json:"version"`
	Listen  int         `json:"listen"`
	Proxy   []ProxyItem `json:"proxy"`
}

func (c *ProxyConfig) String() string {
	data, _ := json.Marshal(c)
	return string(data)
}

func ParseConfig(filePath string) (ProxyConfig, error) {
	var config ProxyConfig

	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return ProxyConfig{}, err
	}
	err = json.Unmarshal(fileData, &config)
	if err != nil {
		return ProxyConfig{}, err
	}
	return config, err
}
