package main

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogap/env_json"
	"github.com/gogap/logs"
)

const INLET_HTTP_API_ENV = "INLET_HTTP_API_ENV"

type InletHTTPAPIConfig struct {
	HTTP               HTTPConfig      `json:"http"`
	Renderer           RendererConfig  `json:"renderer"`
	IncludeConfigFiles []string        `json:"include_config_files"`
	Address            []AddressConfig `json:"address"`
	Graphs             []GraphsConfig  `json:"graphs"`
}

type HTTPConfig struct {
	Address            string            `json:"address"`
	Server             string            `json:"server"`
	APIHeader          string            `json:"api_header"`
	CookiesDomain      string            `json:"cookies_domain"`
	EnableStat         bool              `json:"enable_stat"`
	P3P                string            `json:"p3p"`
	AllowOrigins       []string          `json:"allow_origins"`
	AllowHeaders       []string          `json:"allow_headers"`
	PATH               string            `json:"path"`
	ResponseHeaders    map[string]string `json:"response_headers"`
	PassThroughHeaders []string          `json:"pass_through_headers"`

	_AllowHeaders string          `json:"-"`
	allowOrigins  map[string]bool `json:"-"`
}

type RendererConfig struct {
	DefaultTemplate string              `json:"default_template"`
	Templates       []string            `json:"templates"`
	Variables       []string            `json:"variables"`
	Relation        map[string][]string `json:"relation"`
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
	if p._AllowHeaders != "" {
		return p._AllowHeaders
	}

	if p.AllowHeaders != nil {
		p._AllowHeaders = strings.Join(p.AllowHeaders, ",")
		return p._AllowHeaders
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
		e = fmt.Errorf("read config file of %s failed, error: %s", filename, e)
		panic(e)
	}
	exConf := InletHTTPAPIConfig{}

	envJson := env_json.NewEnvJson(INLET_HTTP_API_ENV, env_json.ENV_JSON_EXT)

	if e = envJson.Unmarshal(bFile, &exConf); e != nil {
		e = fmt.Errorf("unmarshal config file of %s to object failed, error: %s", filename, e)
		panic(e)
	}

	if exConf.Address != nil && len(exConf.Address) > 0 {
		conf.Address = append(conf.Address, exConf.Address...)
	}

	if exConf.Graphs != nil && len(exConf.Graphs) > 0 {
		conf.Graphs = append(conf.Graphs, exConf.Graphs...)
	}

	logs.Info("config file loaded:", filename)

	return
}

func LoadConfig(filename string) InletHTTPAPIConfig {
	bFile, e := ioutil.ReadFile(filename)
	if e != nil {
		panic(e)
	}

	conf := InletHTTPAPIConfig{}

	envJson := env_json.NewEnvJson(INLET_HTTP_API_ENV, env_json.ENV_JSON_EXT)

	if e = envJson.Unmarshal(bFile, &conf); e != nil {
		e = fmt.Errorf("unmarshal config file of %s to object failed, error: %s", filename, e)
		panic(e)
	}

	conf.HTTP.allowOrigins = make(map[string]bool)

	for _, allowOrigin := range conf.HTTP.AllowOrigins {
		conf.HTTP.allowOrigins[allowOrigin] = true
	}

	if conf.HTTP.ResponseHeaders == nil {
		conf.HTTP.ResponseHeaders = make(map[string]string)
	}

	if conf.HTTP.P3P != "" {
		if _, exist := conf.HTTP.ResponseHeaders["P3P"]; !exist {
			conf.HTTP.ResponseHeaders["P3P"] = conf.HTTP.P3P
		}
	}

	if conf.HTTP.Server == "" {
		conf.HTTP.ResponseHeaders["Server"] = conf.HTTP.Server
	} else {
		conf.HTTP.ResponseHeaders["Server"] = "spirit"
	}

	logs.Info("config file loaded:", filename)

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
						filename = strings.TrimRight(filename, "/")
						if filepath.Ext(name) == ".conf" {
							loadIncludeFile(filename+"/"+name, &conf)
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

	internalAllowHeaders := []string{
		"Origin",
		"Content-Type",
		"Authorization",
		"Accept",
		"X-Requested-With",
		"X-Api",
		"X-Api-Multi-Call",
		"X-Api-Call-Timeout"}

	if conf.HTTP.APIHeader != "" {
		internalAllowHeaders = append(internalAllowHeaders, conf.HTTP.APIHeader)
	}

	distinctCache := map[string]string{}

	for _, header := range internalAllowHeaders {
		distinctCache[strings.ToLower(header)] = header
	}

	for _, header := range conf.HTTP.AllowHeaders {
		if _, exist := distinctCache[strings.ToLower(header)]; !exist {
			distinctCache[strings.ToLower(header)] = header
		}
	}

	allowHeaders := []string{}

	for _, header := range distinctCache {
		allowHeaders = append(allowHeaders, header)
	}

	conf.HTTP.AllowHeaders = allowHeaders

	return conf
}
