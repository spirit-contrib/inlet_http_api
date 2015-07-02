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

	ERR_API_RESPONSE_REDNER_FAILED = errors.TN(INLET_HTTP_API_ERR_NS, 12, "render api response error, error: {{.err}}")
	ERR_NO_VALUE_IN_API_RENDER     = errors.TN(INLET_HTTP_API_ERR_NS, 13, "api response have <no value> while render, raw string is: {{.raw}}")

	ERR_ADD_TEMPLATE_FAILED    = errors.TN(INLET_HTTP_API_ERR_NS, 15, "add template failed: {{.file}}, error: {{.err}}")
	ERR_READ_FILE_ERROR        = errors.TN(INLET_HTTP_API_ERR_NS, 16, "read file error, error: {{.err}}")
	ERR_MATCH_FILE_LIST_FAILED = errors.TN(INLET_HTTP_API_ERR_NS, 17, "match file list failed, error: {{.err}}")
	ERR_GET_FILE_INFO          = errors.TN(INLET_HTTP_API_ERR_NS, 18, "get file info failed, error: {{.err}}")
	ERR_DECODE_TMPL_VARS       = errors.TN(INLET_HTTP_API_ERR_NS, 19, "decode template vars error, error: {{.err}}")
	ERR_TMPL_VAR_ALREADY_EXIST = errors.TN(INLET_HTTP_API_ERR_NS, 20, "template vars already exist, key: {{.key}}, value: {{.value}}")
	ERR_TEMPLATE_NOT_EXIST     = errors.TN(INLET_HTTP_API_ERR_NS, 21, "template not exist, name: {{.name}}")
	ERR_API_ALREADY_RELATED    = errors.TN(INLET_HTTP_API_ERR_NS, 22, "api {{.apiName}} already with template {{.tmplName}}")
)
