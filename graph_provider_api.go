package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gogap/errors"
	"github.com/gogap/spirit"
	"github.com/spirit-contrib/inlet_http"
)

const (
	API_HEADER  = "X-Api"
	METHOD_POST = "POST"

	API_RANGE = "X-Range"

	MULTI_CALL       = "X-Api-Multi-Call"
	API_CALL_TIMEOUT = "X-Api-Call-Timeout"
)

type APIGraphProvider struct {
	APIHeader string
	Path      string

	apiGraph map[string]spirit.MessageGraph
}

func NewAPIGraphProvider(apiHeader string, path string, addressConf []AddressConfig, graphConf []GraphsConfig, hooks GraphHooks) inlet_http.GraphProvider {
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

	apiGraph := make(map[string]spirit.MessageGraph)

	for _, graph := range graphConf {
		if apiName, exist := apiGraph[graph.API]; exist {
			panic(fmt.Sprintf("api address already exist,api name: %s", apiName))
		} else {
			addrs := []spirit.MessageAddress{}
			for _, addrName := range graph.Graph {
				if addr, exist := mapAddr[addrName]; exist {
					addrs = append(addrs, addr)
				} else {
					panic(fmt.Sprintf("address of %s not exist", addrName))
				}
			}

			g := make(spirit.MessageGraph)

			g.AddAddress(hooks.Before...)
			g.AddAddress(addrs...)
			g.AddAddress(hooks.After...)

			graph.ErrorAddressName = strings.TrimSpace(graph.ErrorAddressName)
			if graph.ErrorAddressName != "" {
				if addr, exist := mapAddr[graph.ErrorAddressName]; exist {
					g.SetErrorAddress(addr)
				}
			}

			apiGraph[graph.API] = g
		}
	}

	apiHeader = strings.TrimSpace(apiHeader)
	if apiHeader == "" {
		apiHeader = API_HEADER
	}

	return &APIGraphProvider{
		APIHeader: apiHeader,
		apiGraph:  apiGraph,
		Path:      path,
	}
}

func (p *APIGraphProvider) SetGraph(apiName string, graph spirit.MessageGraph) inlet_http.GraphProvider {
	p.apiGraph[apiName] = graph
	return p
}

func (p *APIGraphProvider) GetGraph(r *http.Request, body []byte) (graphs map[string]spirit.MessageGraph, err error) {
	if r.Method != METHOD_POST {
		err = ERR_METHOD_IS_NOT_POST.New(errors.Params{"method": r.Method})
		return
	}

	apiGraphs := map[string]spirit.MessageGraph{}

	appendFunc := func(apiName string) (err error) {
		apiName = strings.TrimSpace(apiName)

		if apiName == "" {
			err = ERR_API_NAME_IS_EMPTY.New()
			return
		}

		if apiGraph, exist := p.apiGraph[apiName]; !exist {
			err = ERR_API_GRAPH_IS_NOT_EXIST.New(errors.Params{"api": apiName})
			return
		} else {
			apiGraphs[apiName] = apiGraph
		}
		return
	}

	if r.Header.Get(MULTI_CALL) == "1" {
		apiParams := map[string]interface{}{}
		if e := json.Unmarshal(body, &apiParams); e != nil {
			err = ERR_UNMARSHAL_MULTI_REQUEST_BODY_FAILED.New(errors.Params{"err": e})
		} else if len(apiParams) > 0 {
			for apiName, _ := range apiParams {
				if err = appendFunc(apiName); err != nil {
					return
				}
			}
		} else {
			err = ERR_EMPTY_MULTI_API_REQUEST.New()
			return
		}
	} else {
		apiName := r.Header.Get(p.APIHeader)
		apiName = strings.TrimSpace(apiName)

		if apiName == "" {
			if p.Path != r.RequestURI {
				apiName = strings.TrimPrefix(r.RequestURI, p.Path+"/")
			}
		}

		if err = appendFunc(apiName); err != nil {
			return
		}
	}

	graphs = apiGraphs

	return
}
