package main

import (
	"github.com/gogap/errors"
)

var INLET_HTTP_API_ERR_NS = "INLET_API"

var (
	ERR_API_NAME_IS_EMPTY              = errors.TN(INLET_HTTP_API_ERR_NS, 1, "api name is empty")
	ERR_METHOD_IS_NOT_POST             = errors.TN(INLET_HTTP_API_ERR_NS, 2, "method is not post, METHOD: {{.method}}")
	ERR_API_GRAPH_IS_NOT_EXIST         = errors.TN(INLET_HTTP_API_ERR_NS, 3, "api graph is not exist, api: {{.api}}")
	ERR_MARSHAL_STRUCT_ERROR           = errors.TN(INLET_HTTP_API_ERR_NS, 4, "marshal struct error, error: {{.err}}")
	ERR_API_REQUEST_TIMEOUT            = errors.TN(INLET_HTTP_API_ERR_NS, 5, "api request timeout, api: {{.api}}")
	ERR_UNMARSHAL_MULTI_REQUEST_FAILED = errors.TN(INLET_HTTP_API_ERR_NS, 6, "unmarshal multi request failed {{.api}}, error: {{.err}}")
	ERR_MULTI_API_REQUEST_NOT_EXIST    = errors.TN(INLET_HTTP_API_ERR_NS, 7, "api not exist in multi request, api: {{.api}}")

	ERR_UNMARSHAL_MULTI_REQUEST_BODY_FAILED = errors.TN(INLET_HTTP_API_ERR_NS, 8, "unmarshal multi request body failed, error: {{.err}}")
	ERR_EMPTY_MULTI_API_REQUEST             = errors.TN(INLET_HTTP_API_ERR_NS, 9, "empty multi api request")

	ERR_PARSE_PROXY_PAYLOAD_FIALED = errors.TN(INLET_HTTP_API_ERR_NS, 10, "parse proxy payload failed, api: {{.api}} error: {{.err}}")

	ERR_PAYLOAD_RESPONSE_COUNT_NOT_MATCH = errors.TN(INLET_HTTP_API_ERR_NS, 11, "payload response count not match")
)
