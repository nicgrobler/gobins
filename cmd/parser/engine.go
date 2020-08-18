package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"text/template"
)

func replace(input, from, to string) string {
	return strings.Replace(input, from, to, -1)
}

func upper(input string) string {
	return strings.ToUpper(input)
}

func lower(input string) string {
	return strings.ToLower(input)
}

func getInterfaceFromTemplate(tpl *template.Template, data interface{}) (result interface{}) {

	b := bytes.Buffer{}
	if err := tpl.Execute(&b, data); err != nil {
		exitLog("program exited due to error in executing template: " + err.Error())
	}

	err := json.Unmarshal(b.Bytes(), &result)
	if err != nil {
		exitLog("program exited due to error in json unmarshalling: " + err.Error())
	}
	return
}

func getTemplateFromFile(fileName, filePath string, funcMap template.FuncMap) *template.Template {
	tpl := template.New(fileName).Funcs(funcMap)
	t, err := tpl.ParseFiles(filePath)
	if err != nil {
		exitLog("program exited due to error in reading template from file " + filePath + ": " + err.Error())
	}
	return t
}

func getTemplateFromString(name, b string, funcMap template.FuncMap) *template.Template {
	tpl := template.New(name).Funcs(funcMap)
	t, err := tpl.Parse(b)
	if err != nil {
		exitLog("program exited due to error in reading template from string " + b + ": " + err.Error())
	}
	return t
}
