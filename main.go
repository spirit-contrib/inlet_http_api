package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gogap/errors"
	"github.com/gogap/logs"
	"github.com/gogap/spirit"
	"github.com/spirit-contrib/inlet_http"
)

const (
	SPIRIT_NAME    = "inlet_http_api"
	METHOD_OPTIONS = "OPTIONS"
)

var (
	conf InletHTTPAPIConfig

	proxyAPI = make(map[string]bool)
)

func main() {
	conf = LoadConfig("./conf/inlet_http_api.conf")

	graphProvider := NewAPIGraphProvider(API_HEADER, conf.Address, conf.Graphs)

	httpConf := inlet_http.Config{Address: conf.HTTP.Address, Domain: conf.HTTP.CookiesDomain}

	inletHTTP := inlet_http.NewInletHTTP(
		inlet_http.SetHTTPConfig(httpConf),
		inlet_http.SetGraphProvider(graphProvider),
		inlet_http.SetResponseHandler(responseHandle),
		inlet_http.SetErrorResponseHandler(errorResponseHandler),
		inlet_http.SetRequestDecoder(requestDecoder),
		inlet_http.SetRequestPayloadHook(requestPayloadHook),
		inlet_http.SetTimeoutHeader(API_CALL_TIMEOUT))

	httpAPISpirit := spirit.NewClassicSpirit(SPIRIT_NAME, "an http inlet with POST request", "1.0.0")
	httpAPIComponent := spirit.NewBaseComponent(SPIRIT_NAME)

	httpAPIComponent.RegisterHandler("callback", inletHTTP.CallBack)
	httpAPIComponent.RegisterHandler("error", inletHTTP.Error)

	httpAPISpirit.Hosting(httpAPIComponent).Build()

	inletHTTP.Requester().SetMessageSenderFactory(httpAPISpirit.GetMessageSenderFactory())

	go inletHTTP.Run(optionHandle)
	httpAPISpirit.Run()
}

type APIResponse struct {
	Code           uint64      `json:"code"`
	ErrorId        string      `json:"error_id,omitempty"`
	ErrorNamespace string      `json:"error_namespace,omitempty"`
	Message        string      `json:"message"`
	Result         interface{} `json:"result"`
}

func requestDecoder(data []byte) (ret map[string]interface{}, err error) {
	str := strings.TrimSpace(string(data))
	if str != "" {
		ret = make(map[string]interface{})
		err = json.Unmarshal(data, &ret)
	}
	return
}

func requestPayloadHook(r *http.Request, apiName string, body []byte, payload *spirit.Payload) (err error) {
	if r.Header.Get(MULTI_CALL) == "1" {
		multiAPIReq := map[string]interface{}{}
		if e := json.Unmarshal(body, &multiAPIReq); e != nil {
			err = ERR_UNMARSHAL_MULTI_REQUEST_FAILED.New(errors.Params{"err": e, "api": apiName})
			return
		} else if reqContent, exist := multiAPIReq[apiName]; exist {
			payload.SetContent(reqContent)
		} else {
			err = ERR_MULTI_API_REQUEST_NOT_EXIST.New(errors.Params{"api": apiName})
			return
		}
	}

	if apiName == "" {
		err = ERR_API_NAME_IS_EMPTY.New()
		return
	}

	if proxyAPI != nil {
		if isProxy, _ := proxyAPI[apiName]; isProxy {
			newPayload := spirit.Payload{}

			if e := newPayload.UnSerialize(body); e != nil {
				err = ERR_PARSE_PROXY_PAYLOAD_FIALED.New(errors.Params{"api": apiName, "err": e})
				logs.Error(err)
				return
			} else {
				payload.CopyFrom(&newPayload)
			}
		}
	}

	payload.SetContext(conf.HTTP.APIHeader, apiName)

	return
}

func optionHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method == METHOD_OPTIONS {
		writeAccessHeaders(w, r)
		writeBasicHeaders(w, r)
		w.Write([]byte(""))
	}
}

func errorResponseHandler(err error, w http.ResponseWriter, r *http.Request) {
	statusCode := http.StatusInternalServerError

	if ERR_API_GRAPH_IS_NOT_EXIST.IsEqual(err) {
		statusCode = http.StatusNotFound
	} else if inlet_http.ERR_REQUEST_TIMEOUT.IsEqual(err) {
		statusCode = http.StatusRequestTimeout
		apiName := r.Header.Get(conf.HTTP.APIHeader)
		err = ERR_API_REQUEST_TIMEOUT.New(errors.Params{"api": apiName})
	}

	var resp APIResponse
	if errCode, ok := err.(errors.ErrCode); ok {
		resp = APIResponse{
			Code:           errCode.Code(),
			ErrorId:        errCode.Id(),
			ErrorNamespace: errCode.Namespace(),
			Message:        errCode.Error(),
			Result:         nil,
		}
	} else {
		resp = APIResponse{
			Code:           500,
			ErrorId:        "",
			ErrorNamespace: INLET_HTTP_API_ERR_NS,
			Message:        err.Error(),
			Result:         nil,
		}
	}

	writeResponseWithStatusCode(&resp, w, r, statusCode)
}

func responseHandle(payloads map[string]spirit.Payload, errs map[string]error, w http.ResponseWriter, r *http.Request) {
	//TODO: improve handle logic
	//X-X-API-MULTI-CALL PROCESS
	if r.Header.Get(MULTI_CALL) == "1" {
		multiResp := map[string]APIResponse{}
		for apiName, payload := range payloads {
			if payload.IsCorrect() {
				multiResp[apiName] = APIResponse{
					Code:   payload.Error().Code,
					Result: payload.GetContent(),
				}
			} else {
				multiResp[apiName] = APIResponse{
					Code:           payload.Error().Code,
					ErrorId:        payload.Error().Id,
					ErrorNamespace: payload.Error().Namespace,
					Message:        payload.Error().Message,
					Result:         nil,
				}
			}
		}

		for apiName, err := range errs {
			if errCode, ok := err.(errors.ErrCode); ok {
				multiResp[apiName] = APIResponse{
					Code:           errCode.Code(),
					ErrorId:        errCode.Id(),
					ErrorNamespace: errCode.Namespace(),
					Message:        errCode.Error(),
					Result:         nil,
				}
			} else {
				multiResp[apiName] = APIResponse{
					Code:           500,
					ErrorId:        "",
					ErrorNamespace: INLET_HTTP_API_ERR_NS,
					Message:        err.Error(),
					Result:         nil,
				}
			}
		}
		resp := APIResponse{
			Code:   0,
			Result: multiResp,
		}
		writeResponse(&resp, w, r)
		return
	}

	lenPayload := len(payloads)
	lenErr := len(errs)

	var resp APIResponse
	if lenPayload+lenErr == 1 {
		if lenPayload == 1 {
			for _, payload := range payloads {
				if payload.IsCorrect() {
					resp = APIResponse{
						Code:   payload.Error().Code,
						Result: payload.GetContent(),
					}
				} else {
					resp = APIResponse{
						Code:           payload.Error().Code,
						ErrorId:        payload.Error().Id,
						ErrorNamespace: payload.Error().Namespace,
						Message:        payload.Error().Message,
						Result:         nil,
					}
				}
			}
		} else if lenErr == 1 {
			for _, err := range errs {
				if errCode, ok := err.(errors.ErrCode); ok {
					resp = APIResponse{
						Code:           errCode.Code(),
						ErrorId:        errCode.Id(),
						ErrorNamespace: errCode.Namespace(),
						Message:        errCode.Error(),
						Result:         nil,
					}
				} else {
					resp = APIResponse{
						Code:           500,
						ErrorId:        "",
						ErrorNamespace: INLET_HTTP_API_ERR_NS,
						Message:        err.Error(),
						Result:         nil,
					}
				}
			}
		}
	} else {
		err := ERR_PAYLOAD_RESPONSE_COUNT_NOT_MATCH.New()
		if errCode, ok := err.(errors.ErrCode); ok {
			resp = APIResponse{
				Code:           errCode.Code(),
				ErrorId:        errCode.Id(),
				ErrorNamespace: errCode.Namespace(),
				Message:        errCode.Error(),
				Result:         nil,
			}
		} else {
			resp = APIResponse{
				Code:           500,
				ErrorId:        "",
				ErrorNamespace: INLET_HTTP_API_ERR_NS,
				Message:        err.Error(),
				Result:         nil,
			}
		}
	}
	writeResponse(&resp, w, r)
}

func writeResponse(v interface{}, w http.ResponseWriter, r *http.Request) {
	writeResponseWithStatusCode(v, w, r, http.StatusOK)
}

func writeResponseWithStatusCode(v interface{}, w http.ResponseWriter, r *http.Request, code int) {
	if data, e := json.Marshal(v); e != nil {
		err := ERR_MARSHAL_STRUCT_ERROR.New(errors.Params{"err": e})
		logs.Error(err)
		if _, ok := v.(error); !ok {
			writeResponseWithStatusCode(&err, w, r, code)
		}
	} else {
		writeAccessHeaders(w, r)
		writeBasicHeaders(w, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		w.Write(data)
	}
}

func writeAccessHeaders(w http.ResponseWriter, r *http.Request) {
	refer := r.Referer()
	if refer == "" {
		refer = r.Header.Get("Origin")
	}

	if refProtocol, refDomain, isAllowd := conf.HTTP.ParseOrigin(refer); isAllowd {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		origin := refProtocol + "://" + refDomain
		if origin == "://" ||
			refProtocol == "chrome-extension" { //issue of post man, chrome limit.
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}

	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", conf.HTTP.allowHeaders())
}

func writeBasicHeaders(w http.ResponseWriter, r *http.Request) {
	for key, value := range conf.HTTP.responseHeaders {
		w.Header().Set(key, value)
	}
}
