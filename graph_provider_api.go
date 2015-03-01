package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gogap/errors"
	"github.com/gogap/spirit"
	"github.com/spirit-contrib/inlet_http"
)

const (
	API_HEADER  = "X-API"
	METHOD_POST = "POST"
)

type APIGraphProvider struct {
	APIHeader string

	apiGraph map[string][]spirit.MessageAddress
}

func NewAPIGraphProvider(apiHeader string, addressConf []AddressConfig, graphConf []GraphsConfig) inlet_http.GraphProvider {
	mapAddr := make(map[string]spirit.MessageAddress)
	for _, addr := range addressConf {
		addr.Name = strings.TrimSpace(addr.Name)
		if addr.Name == "" {
			panic("address name could not be empty")
		}
		if _, exist := mapAddr[addr.Name]; exist {
			panic("address already exist, name: " + addr.Name)
		}
		addr.Url = strings.TrimSpace(addr.Url)
		if addr.Url == "" {
			panic("address url is empty, name: " + addr.Name)
		}
		mapAddr[addr.Name] = spirit.MessageAddress{Type: addr.Type, Url: addr.Url}
	}

	apiGraph := make(map[string][]spirit.MessageAddress)

	for _, graph := range graphConf {
		if apiName, exist := apiGraph[graph.API]; exist {
			panic(fmt.Sprintf("api address already exist,api name: ", apiName))
		} else {
			addrs := []spirit.MessageAddress{}
			for _, addrName := range graph.Graph {
				if addr, exist := mapAddr[addrName]; exist {
					addrs = append(addrs, addr)
				} else {
					panic(fmt.Sprintf("address of %s not exist", addrName))
				}
			}
			apiGraph[graph.API] = addrs
		}
	}

	apiHeader = strings.TrimSpace(apiHeader)
	if apiHeader == "" {
		apiHeader = API_HEADER
	}

	return &APIGraphProvider{
		APIHeader: apiHeader,
		apiGraph:  apiGraph,
	}
}

func (p *APIGraphProvider) SetGraph(apiName string, addr []spirit.MessageAddress) inlet_http.GraphProvider {
	p.apiGraph[apiName] = addr
	return p
}

func (p *APIGraphProvider) GetGraph(r *http.Request) (address []spirit.MessageAddress, err error) {
	if r.Method != METHOD_POST {
		err = ERR_METHOD_IS_NOT_POST.New(errors.Params{"method": r.Method})
		return
	}

	apiName := r.Header.Get(p.APIHeader)
	apiName = strings.TrimSpace(apiName)

	if apiName == "" {
		err = ERR_API_NAME_IS_EMPTY.New()
		return
	}

	if addr, exist := p.apiGraph[apiName]; !exist {
		err = ERR_API_GRAPH_IS_NOT_EXIST.New(errors.Params{"api": apiName})
		return
	} else {
		address = addr
	}
	return
}
