package api_client

import (
	"github.com/gogap/spirit"
)

type APIClient interface {
	Call(apiName string, payload spirit.Payload, v interface{}) (err error)
	Cast(apiName string, payload spirit.Payload) (err error)
}
