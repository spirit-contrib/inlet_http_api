package main

func defaultAPITemplate() (name, tmpl string) {
	name = "default"
	tmpl = `
{
	"code":{{.API.Response.Code}},
	{{if ne .API.Response.ErrorId ""}}"error_id":"{{.API.Response.ErrorId}}",{{end}}
	{{if ne .API.Response.ErrorNamespace ""}}"error_namespace":"{{.API.Response.ErrorNamespace}}",{{end}}
	"message":"{{.API.Response.Message}}",
	{{if .API.IsMulti}}
	"result":{{if .API.Response.Result}}
				{{$outputArray:=newArray}}
				{{range $apiName, $output := .API.Response.Result}}
					{{$out:=printf "\"%s\":%s" $apiName $output}}
					{{$outputArray:=$outputArray.Append $out}}
				{{end}}
				{{$outStr:=$outputArray.Join ","}}
				{{printf "%s%s%s" "{" $outStr "}"}}
			 {{else}}
			 	null
			 {{end}}
	{{else}}
		"result":{{if .API.Response.Result}}{{.API.Response.Result | getJSON}}{{else}}null{{end}}
	{{end}}
}
`
	return
}
