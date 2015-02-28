package main

import (
	"net/http"

	"github.com/gogap/spirit"
	"github.com/spirit-contrib/inlet_http"
)

var (
	requester inlet_http.Requester
)

const (
	SPIRIT_NAME = "inlet_http_api"
)

func main() {
	requester = inlet_http.NewClassicRequester()
	graphProvider := NewAPIGraphProvider()
	graphProvider.SetGraph("rijin.investor.account.register", nil)

	inletHTTP := inlet_http.NewInletHTTP(
		inlet_http.Config{Address: "127.0.0.1:8080"},
		requester,
		graphProvider,
		responseHandler)

	httpAPISpirit := spirit.NewClassicSpirit(SPIRIT_NAME, "an http inlet with POST request", "1.0.0")
	httpAPIComponent := spirit.NewBaseComponent(SPIRIT_NAME)

	httpAPIComponent.RegisterHandler("callback", inletHTTP.MessageResponse)

	httpAPISpirit.Hosting(httpAPIComponent)

	httpAPISpirit.Build()

	receivers := httpAPISpirit.GetComponent(SPIRIT_NAME).GetReceivers("port.callback")

	if receivers == nil {
		panic("inlet_http_api receiver not exist")
	}

	if len(receivers) != 1 {
		panic("the port callback with inlet_http_api should only have one receiver")
	}

	callbackAddr := receivers[0].Address()

	requester.SetMessageSenderFactory(httpAPISpirit.GetMessageSenderFactory())
	requester.Init(callbackAddr)

	go inletHTTP.Run()
	httpAPISpirit.Run()
}

func responseHandler(payload spirit.Payload, w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(payload.Id()))
}
