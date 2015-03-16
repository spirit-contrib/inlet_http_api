package api_client

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gogap/errors"
	"github.com/gogap/spirit"
	"github.com/parnurzeal/gorequest"
)

type HTTPAPIClient struct {
	apiHeaderName string
	timeout       time.Duration
	url           string
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

	apiClient := HTTPAPIClient{
		apiHeaderName: apiHeaderName,
		timeout:       timeout,
		url:           url,
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

	_, body, errs := gorequest.New().Post(p.url).Set(p.apiHeaderName, apiName).Send(string(data)).End()

	var tmpResp struct {
		Code           uint64      `json:"code"`
		ErrorId        string      `json:"error_id,omitempty"`
		ErrorNamespace string      `json:"error_namespace,omitempty"`
		Message        string      `json:"message"`
		Result         interface{} `json:"result"`
	}

	err = errs_to_error(errs)

	if err != nil {
		err = ERR_API_CLIENT_SEND_FAILED.New(errors.Params{"api": apiName, "url": p.url})
		return
	}

	if v != nil {
		tmpResp.Result = v
	}

	if e := json.Unmarshal([]byte(body), &tmpResp); e != nil {
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

func errs_to_error(errs []error) error {
	if errs == nil || len(errs) == 0 {
		return nil
	}

	strErr := ""
	for _, e := range errs {
		strErr += e.Error() + "; "
	}
	return errors.New(strErr)
}
