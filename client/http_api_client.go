package api_client

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gogap/errors"
	"github.com/gogap/spirit"
	"github.com/mreiferson/go-httpclient"
)

var (
	DefaultClientTimeout = time.Second * 5
)

type HTTPAPIClient struct {
	apiHeaderName string
	url           string
	client        *http.Client
}

func NewHTTPAPIClient(url string, apiHeaderName string, timeout time.Duration) APIClient {
	url = strings.TrimSpace(url)
	apiHeaderName = strings.TrimSpace(apiHeaderName)

	if url == "" {
		panic("url could not be nil")
	}

	if apiHeaderName == "" {
		apiHeaderName = "X-API"
	}

	if timeout <= 0 {
		timeout = DefaultClientTimeout
	}

	transport := &httpclient.Transport{
		ConnectTimeout:        timeout,
		RequestTimeout:        timeout,
		ResponseHeaderTimeout: timeout,
	}

	apiClient := HTTPAPIClient{
		apiHeaderName: apiHeaderName,
		url:           url,
		client:        &http.Client{Transport: transport},
	}
	return &apiClient
}

func (p *HTTPAPIClient) Call(apiName string, payload spirit.Payload, v interface{}) (err error) {
	apiName = strings.TrimSpace(apiName)

	if apiName == "" {
		err = ERR_API_NAME_IS_EMPTY.New()
		return
	}

	var data []byte
	if data, err = payload.Serialize(); err != nil {
		return
	}

	postBodyReader := bytes.NewReader(data)

	var req *http.Request
	if req, err = http.NewRequest("POST", p.url, postBodyReader); err != nil {
		err = ERR_API_CLIENT_CREATE_NEW_REQUEST_FAILED.New(errors.Params{"err": err})
		return
	}

	req.Header.Add(p.apiHeaderName, apiName)

	var resp *http.Response
	if resp, err = p.client.Do(req); err != nil {
		err = ERR_API_CLIENT_SEND_FAILED.New(errors.Params{"api": apiName, "url": p.url})
		return
	}

	var body []byte

	if resp != nil {
		defer resp.Body.Close()

		if bBody, e := ioutil.ReadAll(resp.Body); e != nil {
			err = ERR_API_CLIENT_READ_RESPONSE_BODY_FAILED.New(errors.Params{"api": apiName, "err": e})
			return
		} else if resp.StatusCode != http.StatusOK {
			err = ERR_API_CLIENT_BAD_STATUS_CODE.New(errors.Params{"api": apiName, "code": resp.StatusCode})
			return
		} else {
			body = bBody
		}
	}

	var tmpResp struct {
		Code           uint64      `json:"code"`
		ErrorId        string      `json:"error_id,omitempty"`
		ErrorNamespace string      `json:"error_namespace,omitempty"`
		Message        string      `json:"message"`
		Result         interface{} `json:"result"`
	}

	if v != nil {
		tmpResp.Result = v
	}

	if e := json.Unmarshal(body, &tmpResp); e != nil {
		err = ERR_API_CLIENT_RESPONSE_UNMARSHAL_FAILED.New(errors.Params{"api": apiName, "url": p.url, "err": e})
		return
	}

	if tmpResp.Code == 0 {
		return
	} else {
		err = errors.NewErrorCode(tmpResp.ErrorId, tmpResp.Code, tmpResp.ErrorNamespace, tmpResp.Message, "", nil)
		return
	}

	return
}

func (p *HTTPAPIClient) Cast(apiName string, payload spirit.Payload) (err error) {
	return p.Call(apiName, payload, nil)
}
