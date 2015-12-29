package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/go-martini/martini"
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

var (
	responseRenderer *APIResponseRenderer
)

var (
	renderedXDomainProxy string
)

func main() {
	logs.SetFileLogger("logs/inlet_http_api.log")

	httpAPISpirit := spirit.NewClassicSpirit(
		SPIRIT_NAME,
		"an http inlet with POST request",
		"1.0.0",
		[]spirit.Author{
			{Name: "zeal", Email: "xujinzheng@gmail.com"},
		},
	)

	httpAPIComponent := spirit.NewBaseComponent(SPIRIT_NAME)

	inletHTTP := inlet_http.NewInletHTTP()

	httpAPIComponent.RegisterHandler("callback", inletHTTP.CallBack)
	httpAPIComponent.RegisterHandler("error", inletHTTP.Error)

	funcStartInletHTTP := func() error {
		conf = LoadConfig("conf/inlet_http_api.conf")

		graphProvider := NewAPIGraphProvider(API_HEADER, conf.HTTP.PATH, conf.Address, conf.Graphs)

		httpConf := inlet_http.Config{
			Address:    conf.HTTP.Address,
			Domain:     conf.HTTP.CookiesDomain,
			EnableStat: conf.HTTP.EnableStat,
		}

		emptyLogger := log.New(new(EmptyWriter), "", 0)

		inletHTTP.Option(inlet_http.SetHTTPConfig(httpConf),
			inlet_http.SetGraphProvider(graphProvider),
			inlet_http.SetResponseHandler(responseHandle),
			inlet_http.SetErrorResponseHandler(errorResponseHandler),
			inlet_http.SetRequestDecoder(requestDecoder),
			inlet_http.SetRequestPayloadHook(requestPayloadHook),
			inlet_http.SetTimeoutHeader(API_CALL_TIMEOUT),
			inlet_http.SetRangeHeader(API_RANGE),
			inlet_http.SetPassThroughHeaders(conf.HTTP.PassThroughHeaders...),
			inlet_http.SetLogger(emptyLogger))

		inletHTTP.Requester().SetMessageSenderFactory(spirit.GetMessageSenderFactory())

		if e := responseRenderer.LoadTemplates(conf.Renderer.Templates...); e != nil {
			panic(e)
		}

		if e := responseRenderer.SetDefaultTemplate(conf.Renderer.DefaultTemplate); e != nil {
			panic(e)
		}

		if e := responseRenderer.LoadVariables(conf.Renderer.Variables...); e != nil {
			panic(e)
		}

		if conf.Renderer.Relation != nil {
			for name, apis := range conf.Renderer.Relation {
				for _, api := range apis {
					if e := responseRenderer.SetAPITemplate(api, name); e != nil {
						panic(e)
					}
				}
			}
		}

		if httpConf.EnableStat {
			inletHTTP.Group(conf.HTTP.PATH, func(r martini.Router) {
				r.Post("", inletHTTP.Handler)
				r.Post("/:apiName", inletHTTP.Handler)
				r.Options("", optionHandle)
				r.Options("/:apiName", optionHandle)
			}, martini.Static("stat"))

		} else {
			inletHTTP.Group(conf.HTTP.PATH, func(r martini.Router) {
				r.Post("", inletHTTP.Handler)
				r.Post("/:apiName", inletHTTP.Handler)
				r.Options("", optionHandle)
				r.Options("/:apiName", optionHandle)
			})
		}

		inletHTTP.Group("/", func(r martini.Router) {
			r.Get("xdomain/proxy.html", func() string {
				return renderedXDomainProxy
			})

			r.Get("xdomain/lib/xdomain.min.js", func() string {
				return xdomainProxyJS()
			})

			r.Get("ping", func() string {
				return "pong"
			})
		})

		renderedXDomainProxy = renderXDomainProxy(conf.HTTP.AllowOrigins, conf.HTTP.PATH)

		go inletHTTP.Run()

		return nil
	}

	responseRenderer = NewAPIResponseRenderer()

	httpAPISpirit.Hosting(httpAPIComponent, funcStartInletHTTP).Build().Run()
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
		decoder := json.NewDecoder(strings.NewReader(str))
		decoder.UseNumber()
		ret = make(map[string]interface{})

		if err = decoder.Decode(&ret); err != nil {
			return
		}
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

	apiName := r.Header.Get(conf.HTTP.APIHeader)
	if text, e := responseRenderer.Render(false, map[string]APIResponse{apiName: resp}); e != nil {
		err := ERR_API_RESPONSE_REDNER_FAILED.New(errors.Params{"err": e})
		eResp := APIResponse{
			Code:           err.Code(),
			ErrorId:        err.Id(),
			ErrorNamespace: err.Namespace(),
			Message:        err.Error(),
			Result:         nil,
		}
		writeResponse(&eResp, w, r)
		return
	} else {
		writeTextResponse(text, w, r)
		return
	}
}

func responseHandle(graphsResponse map[string]inlet_http.GraphResponse, w http.ResponseWriter, r *http.Request) {
	isMultiCall := r.Header.Get(MULTI_CALL) == "1"

	multiResp := map[string]APIResponse{}
	for apiName, graphResponse := range graphsResponse {
		if graphResponse.Error != nil {
			if errCode, ok := graphResponse.Error.(errors.ErrCode); ok {
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
					Message:        graphResponse.Error.Error(),
					Result:         nil,
				}
			}
		} else if graphResponse.RespPayload.IsCorrect() {
			multiResp[apiName] = APIResponse{
				Code:   graphResponse.RespPayload.Error().Code,
				Result: graphResponse.RespPayload.GetContent(),
			}
		} else {
			multiResp[apiName] = APIResponse{
				Code:           graphResponse.RespPayload.Error().Code,
				ErrorId:        graphResponse.RespPayload.Error().Id,
				ErrorNamespace: graphResponse.RespPayload.Error().Namespace,
				Message:        graphResponse.RespPayload.Error().Message,
				Result:         nil,
			}
		}
	}

	if text, e := responseRenderer.Render(isMultiCall, multiResp); e != nil {
		err := ERR_API_RESPONSE_REDNER_FAILED.New(errors.Params{"err": e})
		resp := APIResponse{
			Code:           err.Code(),
			ErrorId:        err.Id(),
			ErrorNamespace: err.Namespace(),
			Message:        err.Error(),
			Result:         nil,
		}
		writeResponse(&resp, w, r)
		return
	} else {
		writeTextResponse(text, w, r)
		return
	}
}

func sha1hash(hash []byte) []byte {
	h := sha1.New()
	h.Write(hash)
	return h.Sum(nil)
}

func signatureResponse(data []byte, w http.ResponseWriter) {
	if !conf.HTTP.Signature.Enabled {
		return
	}

	hash := sha1hash(data)

	if bSignature, err := rsa.SignPKCS1v15(rand.Reader, conf.HTTP.Signature._PrivateKey, crypto.SHA1, sha1hash(hash)); err != nil {
		logs.Error(err)
	} else {
		signature := base64.StdEncoding.EncodeToString(bSignature)
		w.Header().Set(conf.HTTP.Signature.Header, signature)
	}
}

func writeTextResponse(text string, w http.ResponseWriter, r *http.Request) {
	writeAccessHeaders(w, r)
	writeBasicHeaders(w, r)
	signatureResponse([]byte(text), w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(text))
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
		signatureResponse(data, w)
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
	for key, value := range conf.HTTP.ResponseHeaders {
		w.Header().Set(key, value)
	}
}
