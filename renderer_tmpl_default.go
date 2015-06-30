package main

func defaultAPITemplate() (name, tmpl string) {
	name = "default"
	tmpl = `
{
	"code":{{.API.Response.Code}},
	{{if ne .API.Response.ErrorId ""}}"error_id":"{{.API.Response.ErrorId}}",{{end}}
	{{if ne .API.Response.ErrorNamespace ""}}"error_namespace":"{{.API.Response.ErrorNamespace}}",{{end}}
	"message":"{{.API.Response.Message}}",
	"result":{{getJSON .API.Response.Result}}
}
`
	return
}
