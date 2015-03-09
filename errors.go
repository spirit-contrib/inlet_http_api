package main

import (
	"github.com/gogap/errors"
)

var INLET_HTTP_API_ERR_NS = "INLET_API"

var (
	ERR_API_NAME_IS_EMPTY      = errors.TN(INLET_HTTP_API_ERR_NS, 1, "api name is empty")
	ERR_METHOD_IS_NOT_POST     = errors.TN(INLET_HTTP_API_ERR_NS, 2, "method is not post, METHOD: {{.method}}")
	ERR_API_GRAPH_IS_NOT_EXIST = errors.TN(INLET_HTTP_API_ERR_NS, 3, "api graph is not exist, api: {{.api}}")
	ERR_MARSHAL_STRUCT_ERROR   = errors.TN(INLET_HTTP_API_ERR_NS, 4, "marshal struct error, error: {{.err}}")
	ERR_API_REQUEST_TIMEOUT    = errors.TN(INLET_HTTP_API_ERR_NS, 5, "api request timeout, api: {{.api}}")
)
