package main

import (
	"encoding/json"
	"net/http"

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
		inlet_http.SetRequestDecoder(requestDecoder))

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
	ret = make(map[string]interface{})
	err = json.Unmarshal(data, &ret)
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
	w.WriteHeader(http.StatusInternalServerError)
	writeResponse(&resp, w, r)
}

func responseHandle(payload spirit.Payload, w http.ResponseWriter, r *http.Request) {
	if payload.IsCorrect() {
		correctHandle(payload, w, r)
	} else {
		errorHandle(payload, w, r)
	}
}

func correctHandle(payload spirit.Payload, w http.ResponseWriter, r *http.Request) {
	resp := APIResponse{
		Code:   payload.Error().Code,
		Result: payload.GetContent(),
	}
	writeResponse(&resp, w, r)
}

func errorHandle(payload spirit.Payload, w http.ResponseWriter, r *http.Request) {
	resp := APIResponse{
		Code:           payload.Error().Code,
		ErrorId:        payload.Error().Id,
		ErrorNamespace: payload.Error().Namespace,
		Message:        payload.Error().Message,
		Result:         nil,
	}

	w.WriteHeader(http.StatusInternalServerError)
	writeResponse(&resp, w, r)
}

func writeResponse(v interface{}, w http.ResponseWriter, r *http.Request) {
	if data, e := json.Marshal(v); e != nil {
		err := ERR_MARSHAL_STRUCT_ERROR.New(errors.Params{"err": e})
		logs.Error(err)
		if _, ok := v.(error); !ok {
			writeResponse(&err, w, r)
		}
	} else {
		writeAccessHeaders(w, r)
		writeBasicHeaders(w, r)
		w.Header().Set("Content-Type", "application/json")
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
