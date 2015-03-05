package api_client

import (
	"github.com/gogap/errors"
)

var INLET_HTTP_API_CLIENT_ERR_NS = "INLET_API_CLIENT"

var (
	ERR_API_NAME_IS_EMPTY                    = errors.TN(INLET_HTTP_API_CLIENT_ERR_NS, 1, "api name is empty")
	ERR_API_CLIENT_SEND_FAILED               = errors.TN(INLET_HTTP_API_CLIENT_ERR_NS, 2, "api client send failed, api: {{.api}}, url: {{.url}}")
	ERR_API_CLIENT_RESPONSE_UNMARSHAL_FAILED = errors.TN(INLET_HTTP_API_CLIENT_ERR_NS, 3, "api response unmarshal failed, api: {{.api}}, url: {{.url}}, err: {{.err}}")
	ERR_API_CLIENT_ERROR_MESSAGE             = errors.TN(INLET_HTTP_API_CLIENT_ERR_NS, 4, "remote server response error, api: {{.api}}, url: {{.url}} error namespace: {{.err_namespace}}, code: {{.code}}, message: {{.err_namespace}}, error id: {{.err_id}}")
)
