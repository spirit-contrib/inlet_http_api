package main

import (
	"net/http"
	"strings"

	"github.com/gogap/spirit"
	"github.com/spirit-contrib/inlet_http"
)

const (
	DEFAULT_API_HEADER = "X-API"
	METHOD_POST        = "POST"
)

type APIGraphProvider struct {
	APIHeader string

	apiGraph map[string][]spirit.MessageAddress
}

func NewAPIGraphProvider() inlet_http.GraphProvider {
	return &APIGraphProvider{
		APIHeader: DEFAULT_API_HEADER,
		apiGraph:  make(map[string][]spirit.MessageAddress),
	}
}

func (p *APIGraphProvider) SetGraph(apiName string, addr []spirit.MessageAddress) inlet_http.GraphProvider {
	p.apiGraph[apiName] = addr
	return p
}

func (p *APIGraphProvider) GetGraph(r *http.Request) (address []spirit.MessageAddress, err error) {
	if r.Method != METHOD_POST {
		//ERR
		return
	}

	apiName := r.Header.Get(p.APIHeader)
	apiName = strings.TrimSpace(apiName)

	if apiName == "" {
		//ERR
		return
	}

	if addr, exist := p.apiGraph[apiName]; !exist {
		//err
		return
	} else {
		address = addr
	}
	return
}
