package api_client

import (
	"github.com/gogap/errors"
)

var INLET_HTTP_API_CLIENT_ERR_NS = "INLET_API_CLIENT"

var (
	ERR_API_NAME_IS_EMPTY                    = errors.TN(INLET_HTTP_API_CLIENT_ERR_NS, 1, "api name is empty")
	ERR_API_CLIENT_SEND_FAILED               = errors.TN(INLET_HTTP_API_CLIENT_ERR_NS, 2, "api client send failed, api: {{.api}}, url: {{.url}}, err: {{.err}}")
	ERR_API_CLIENT_RESPONSE_UNMARSHAL_FAILED = errors.TN(INLET_HTTP_API_CLIENT_ERR_NS, 3, "api response unmarshal failed, api: {{.api}}, url: {{.url}}, err: {{.err}}")

	ERR_API_CLIENT_READ_RESPONSE_BODY_FAILED = errors.TN(INLET_HTTP_API_CLIENT_ERR_NS, 4, "read api response body failed, api is: {{.api}},err: {{.err}}")
	ERR_API_CLIENT_BAD_STATUS_CODE           = errors.TN(INLET_HTTP_API_CLIENT_ERR_NS, 5, "bad response status code, api is: {{.api}}, code is: {{.code}}")
	ERR_API_CLIENT_CREATE_NEW_REQUEST_FAILED = errors.TN(INLET_HTTP_API_CLIENT_ERR_NS, 6, "create new request failed, err: {{.err}}")
)
