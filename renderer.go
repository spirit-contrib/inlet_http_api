package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gogap/errors"
)

type APIRenderData struct {
	IsMulti  bool
	Name     string
	Response APIResponse
}

type RenderData struct {
	API  APIRenderData
	Vars map[string]interface{}
}

type APIResponseRenderer struct {
	apiTemplate map[string]string
	template.Template
	Variables       map[string]interface{}
	defaultTemplate string
}

func NewAPIResponseRenderer() *APIResponseRenderer {
	render := &APIResponseRenderer{
		Template:        *template.New(""),
		apiTemplate:     make(map[string]string),
		defaultTemplate: "_internal/default",
		Variables:       make(map[string]interface{}),
	}

	render.Funcs(funcMap)

	if e := render.AddInternalTemplate(defaultAPITemplate()); e != nil {
		panic(e)
	}

	return render
}

func (p *APIResponseRenderer) LoadTemplates(paths ...string) (err error) {
	if paths != nil {
		for _, path := range paths {
			if fi, e := os.Stat(path); e != nil {
				err = ERR_GET_FILE_INFO.New(errors.Params{"err": e})
				return
			} else if !fi.IsDir() {
				base := filepath.Base(path)
				if base[0] != '.' && base[0] != '~' {
					if tmplData, e := ioutil.ReadFile(path); e != nil {
						err = ERR_READ_FILE_ERROR.New(errors.Params{"err": e})
						return
					} else {

						if e := p.AddTemplate(base, string(tmplData)); e != nil {
							err = ERR_ADD_TEMPLATE_FAILED.New(errors.Params{"file": path})
							return
						}
					}
				}
			} else {
				if matches, e := filepath.Glob(filepath.Join(path, "*.tmpl")); e != nil {
					err = ERR_MATCH_FILE_LIST_FAILED.New(errors.Params{"err": e})
					return
				} else {

					for _, file := range matches {
						name, _ := filepath.Rel(path, file)
						if name[0] != '.' && name[0] != '~' {
							if tmplData, e := ioutil.ReadFile(file); e != nil {
								err = ERR_READ_FILE_ERROR.New(errors.Params{"err": e})
								return
							} else {

								if e := p.AddTemplate(name, string(tmplData)); e != nil {
									err = ERR_ADD_TEMPLATE_FAILED.New(errors.Params{"file": file, "err": e})
									return
								}
							}
						}
					}
				}
			}
		}
	}

	return
}

func (p *APIResponseRenderer) LoadVariables(paths ...string) (err error) {
	if paths != nil {
		for _, path := range paths {
			if fi, e := os.Stat(path); e != nil {
				err = ERR_GET_FILE_INFO.New(errors.Params{"err": e})
				return
			} else if !fi.IsDir() {
				base := filepath.Base(path)
				if base[0] != '.' && base[0] != '~' {
					if jsonData, e := ioutil.ReadFile(path); e != nil {
						err = ERR_READ_FILE_ERROR.New(errors.Params{"err": e})
						return
					} else {
						decodder := json.NewDecoder(bytes.NewReader(jsonData))
						decodder.UseNumber()
						vars := map[string]interface{}{}
						if e := decodder.Decode(&vars); e != nil {
							err = ERR_DECODE_TMPL_VARS.New(errors.Params{"err": e})
							return
						}
						if err = p.appendVars(vars); err != nil {
							return
						}
					}
				}
			} else {
				if matches, e := filepath.Glob("*.tmpl"); e != nil {
					err = ERR_MATCH_FILE_LIST_FAILED.New(errors.Params{"err": e})
					return
				} else {
					for _, file := range matches {
						name, _ := filepath.Rel(path, file)
						if name[0] != '.' && name[0] != '~' {
							if jsonData, e := ioutil.ReadFile(filepath.Join(path, file)); e != nil {
								err = ERR_READ_FILE_ERROR.New(errors.Params{"err": e})
								return
							} else {
								decodder := json.NewDecoder(bytes.NewReader(jsonData))
								decodder.UseNumber()
								vars := map[string]interface{}{}
								if e := decodder.Decode(&vars); e != nil {
									err = ERR_DECODE_TMPL_VARS.New(errors.Params{"err": e})
									return
								}
								if err = p.appendVars(vars); err != nil {
									return
								}
							}
						}
					}
				}
			}
		}
	}
	return
}

func (p *APIResponseRenderer) SetDefaultTemplate(name string) (err error) {
	name = strings.TrimSpace(name)
	if name == "" {
		p.defaultTemplate = "_internal/default"
		return
	}

	if p.Lookup(name) != nil {
		p.defaultTemplate = name
	} else {
		err = ERR_TEMPLATE_NOT_EXIST.New(errors.Params{"name": name})
		return
	}
	return
}

func (p *APIResponseRenderer) ResetAPITemplate(apiName string) {
	if _, exist := p.apiTemplate[apiName]; exist {
		delete(p.apiTemplate, apiName)
	}
}

func (p *APIResponseRenderer) SetAPITemplate(apiName, tplName string) (err error) {
	if originalName, exist := p.apiTemplate[apiName]; exist {
		if originalName != tplName {
			err = ERR_API_ALREADY_RELATED.New(errors.Params{"apiName": apiName, "tmplName": originalName})
			return
		}
		return
	}

	if p.Lookup(tplName) != nil {
		p.apiTemplate[apiName] = tplName
		return
	} else {
		err = ERR_TEMPLATE_NOT_EXIST.New(errors.Params{"name": tplName})
		return
	}

	return
}

func (p *APIResponseRenderer) AddInternalTemplate(name, tpl string) error {
	return p.AddTemplate("_internal/"+name, tpl)
}

func (p *APIResponseRenderer) AddTemplate(name, tpl string) (err error) {
	tpl = strings.Replace(tpl, "\n", "", -1)
	tpl = strings.Replace(tpl, "\t", "", -1)
	_, err = p.New(name).Parse(tpl)
	return
}

func (p *APIResponseRenderer) Render(isMulti bool, response map[string]APIResponse) (text string, err error) {
	output := map[string]string{}

	for api, response := range response {

		renderData := RenderData{
			API: APIRenderData{
				false,
				api,
				response,
			},
			Vars: p.Variables,
		}

		tmplName := p.defaultTemplate
		if name, exist := p.apiTemplate[api]; exist {
			tmplName = name
		}

		var buf bytes.Buffer
		if err = p.ExecuteTemplate(&buf, tmplName, renderData); err != nil {
			return
		}

		if !isMulti {
			text = buf.String()
			return
		}

		output[api] = buf.String()
	}

	var buf bytes.Buffer

	multiResponse := APIResponse{
		Code:    0,
		Message: "",
		Result:  output,
	}

	multiRenderData := RenderData{
		API: APIRenderData{
			true,
			"",
			multiResponse,
		},
		Vars: p.Variables,
	}

	if err = p.ExecuteTemplate(&buf, p.defaultTemplate, multiRenderData); err != nil {
		return
	}

	text = buf.String()

	return
}

func (p *APIResponseRenderer) appendVars(vars map[string]interface{}) (err error) {
	for k, v := range vars {
		if original, exist := p.Variables[k]; exist {
			if original != v {
				err = ERR_TMPL_VAR_ALREADY_EXIST.New(errors.Params{"key": k, "value": v})
				return
			}
		} else {
			p.Variables[k] = v
		}
	}
	return
}
