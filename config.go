package main

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type InletHTTPAPIConfig struct {
	HTTP               HTTPConfig      `json:"http"`
	IncludeConfigFiles []string        `json:"include_config_files"`
	Address            []AddressConfig `json:"address"`
	Graphs             []GraphsConfig  `json:"graphs"`
}

type HTTPConfig struct {
	Address       string   `json:"address"`
	Server        string   `json:"server"`
	APIHeader     string   `json:"api_header"`
	CookiesDomain string   `json:"cookies_domain"`
	P3P           string   `json:"p3p"`
	AllowOrigins  []string `json:"allow_origins"`
	AllowHeaders  []string `json:"allow_headers"`

	allowOrigins    map[string]bool   `json:"-"`
	responseHeaders map[string]string `json:"-"`
}

type AddressConfig struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Url  string `json:"url"`
}

type GraphsConfig struct {
	API              string   `json:"api"`
	Graph            []string `json:"graph"`
	IsProxy          bool     `json:"is_proxy,omitempty"`
	ErrorAddressName string   `json:"error_address_name"`
}

func parseRefer(url string) (protocol string, domain string) {
	url = strings.TrimSpace(url)

	if len(url) > 0 {
		start0 := strings.Index(url, "://")
		url0 := url[start0+3 : len(url)]
		surls := strings.Split(url0, "/")

		if len(surls) > 0 {
			domain = surls[0]
		}

		protocol = url[0:start0]
	}

	return
}

func (p *HTTPConfig) ParseOrigin(refer string) (protocol string, domin string, isAllow bool) {
	if _, err := url.Parse(refer); err == nil {
		refProtocol, refDomain := parseRefer(refer)
		if p.allowOrigins["*"] ||
			p.allowOrigins[refDomain] {
			return refProtocol, refDomain, true
		}
		return refProtocol, refDomain, false
	}

	return "", "", false
}

func (p *HTTPConfig) allowHeaders() string {
	if p.AllowHeaders != nil {
		return strings.Join(p.AllowHeaders, ",")
	}
	return ""
}

func isFileOrDir(filename string, decideDir bool) bool {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return false
	}
	isDir := fileInfo.IsDir()
	if decideDir {
		return isDir
	}
	return !isDir
}

func loadIncludeFile(filename string, conf *InletHTTPAPIConfig) {

	bFile, e := ioutil.ReadFile(filename)
	if e != nil {
		panic(e)
	}
	exConf := InletHTTPAPIConfig{}
	if e = json.Unmarshal(bFile, &exConf); e != nil {
		panic(e)
	}

	if exConf.Address != nil && len(exConf.Address) > 0 {
		conf.Address = append(conf.Address, exConf.Address...)
	}

	if exConf.Graphs != nil && len(exConf.Graphs) > 0 {
		conf.Graphs = append(conf.Graphs, exConf.Graphs...)
	}

	return
}

func LoadConfig(filename string) InletHTTPAPIConfig {
	bFile, e := ioutil.ReadFile(filename)
	if e != nil {
		panic(e)
	}

	conf := InletHTTPAPIConfig{}
	if e = json.Unmarshal(bFile, &conf); e != nil {
		panic(e)
	}

	conf.HTTP.allowOrigins = make(map[string]bool)

	for _, allowOrigin := range conf.HTTP.AllowOrigins {
		conf.HTTP.allowOrigins[allowOrigin] = true
	}

	if conf.HTTP.responseHeaders == nil {
		conf.HTTP.responseHeaders = make(map[string]string)
	}

	if conf.HTTP.P3P != "" {
		conf.HTTP.responseHeaders["P3P"] = conf.HTTP.P3P
	}

	if conf.HTTP.Server == "" {
		conf.HTTP.responseHeaders["Server"] = conf.HTTP.Server
	} else {
		conf.HTTP.responseHeaders["Server"] = "spirit"
	}

	//read include configs
	if conf.IncludeConfigFiles != nil && len(conf.IncludeConfigFiles) > 0 {
		for _, filename := range conf.IncludeConfigFiles {
			if isFileOrDir(filename, true) {
				if f, e := os.Open(filename); e != nil {
					panic(e)
				} else if names, e := f.Readdirnames(-1); e != nil {
					panic(e)
				} else {
					for _, name := range names {
						if filepath.Ext(name) == "conf" {
							loadIncludeFile(name, &conf)
						}
					}
				}
			} else {
				loadIncludeFile(filename, &conf)
			}
		}
	}

	for _, graph := range conf.Graphs {
		if graph.IsProxy {
			proxyAPI[graph.API] = true
		}
	}

	return conf
}
