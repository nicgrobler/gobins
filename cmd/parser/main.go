package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"
)

func getFuncMap() template.FuncMap {
	return template.FuncMap{
		"replace":    replace,
		"upper":      upper,
		"lower":      lower,
		"getCPU":     getCPU,
		"getMEM":     getMEM,
		"getPVC":     getPVC,
		"getStorage": getStorage,
	}
}

func getCPU(data *expectedInput, defaultValue interface{}) string {
	if o := data.getOptional("cpu"); o != nil {
		// CPU can be specified with, and without a suffix - handle both
		if o.Unit.string != "" {
			return quoteString(concat(o.Count.int, o.Unit.string))
		}
		return strconv.Itoa(o.Count.int)
	}
	// use the supplied default
	switch t := defaultValue.(type) {
	case string:
		return quoteString(t)
	case int:
		return strconv.Itoa(t)
	}
	return ""
}

func getMEM(data *expectedInput, defaultValue interface{}) string {
	if o := data.getOptional("memory"); o != nil {
		return quoteString(concat(o.Count.int, o.Unit.string))
	}
	// use the supplied default
	switch t := defaultValue.(type) {
	case string:
		return quoteString(t)
	case int:
		return strconv.Itoa(t)
	}
	return ""
}

func getPVC(data *expectedInput, defaultValue interface{}) string {
	if o := data.getOptional("volumes"); o != nil {
		return strconv.Itoa(o.Count.int)
	}
	// use the supplied default
	switch t := defaultValue.(type) {
	case string:
		return quoteString(t)
	case int:
		return strconv.Itoa(t)
	}
	return ""
}

func getStorage(data *expectedInput, defaultValue interface{}) string {
	if o := data.getOptional("storage"); o != nil {
		return quoteString(concat(o.Count.int, o.Unit.string))
	}
	// use the supplied default
	switch t := defaultValue.(type) {
	case string:
		return quoteString(t)
	case int:
		return strconv.Itoa(t)
	}
	return ""
}

func quoteString(s string) string {
	return "\"" + s + "\""
}

func (c *config) createJSONBytes(data *expectedInput, tpl *template.Template) ([]byte, error) {

	unknown := getInterfaceFromTemplate(tpl, data)
	if c.flatOutput {
		bytes, err := json.Marshal(unknown)
		if err != nil {
			return nil, err
		}
		return bytes, nil
	}
	bytes, err := json.MarshalIndent(unknown, "", "  ")
	if err != nil {
		return nil, err
	}
	return bytes, nil

}

func concat(i int, s string) string {
	return strconv.Itoa(i) + s
}

func rangeBetweenBrackets(results []byte) []byte {
	// given a slice with '[' as the first char, returns the next, all the way to the second from last, the closing ']'
	if results[0] == byte('[') {
		return results[1 : len(results)-2]
	}
	return results
}

func (c *config) getTemplates(data *expectedInput) []*template.Template {
	var templates []*template.Template
	if c.usefileContentInput {
		tpl := getTemplateFromString("raw_stream", c.fileContent, getFuncMap())
		templates = append(templates, tpl)
		return templates
	}
	for _, fileName := range c.fileList {
		tpl := getTemplateFromFile(fileName, c.templateDir+fileName, getFuncMap())
		templates = append(templates, tpl)
	}
	return templates
}

func (c *config) process(data *expectedInput) ([]byte, error) {

	var results []byte
	// as returning a JSON slice, add first and last brackets
	results = append(results, byte('['))

	// grab templates
	templates := c.getTemplates(data)
	for _, t := range templates {
		tempBytes, err := c.createJSONBytes(data, t)
		if err != nil {
			return nil, err
		}
		for _, b := range rangeBetweenBrackets(tempBytes) {
			results = append(results, b)
		}

	}
	results = append(results, byte(']'))
	return results, nil
}

/*
	Helpers
*/

func logFunction(format string) {
	fmt.Println(format)
	os.Exit(1)
}

var exitLog = logFunction

type config struct {
	usefileContentInput bool
	flatOutput          bool
	templateDir         string
	fileList            []string
	fileContent         string // optional, allows testing, and runtime funkiness if required
}

func stringToSlice(name string) []string {
	val := strings.Split(name, ",")
	return val
}

func getConfig(fileContent string) (*config, error) {
	tdir := os.Getenv("TEMPLATEDIR")
	flist := os.Getenv("TEMPLATE_FILELIST")

	if tdir == "" {
		if flist == "" {
			if fileContent == "" {
				return nil, errors.New("environment variables undefined")
			}
			return &config{usefileContentInput: true, fileContent: fileContent}, nil
		}
		return &config{fileList: stringToSlice(flist)}, nil
	}
	// ensure tdir ends with a "/"
	if tdir[len(tdir)-1] != '/' {
		tdir = tdir + "/"
	}
	if flist == "" {
		return nil, errors.New("environment variables undefined")
	}
	return &config{templateDir: tdir, fileList: stringToSlice(flist)}, nil

}

func main() {

	var incomingJSON *string
	incomingJSON = flag.String("generate", "", "the json payload used to generate the OpenShift json")
	flag.Parse()

	if *incomingJSON == "" {
		exitLog("program exited due to missing input")
	}

	var inputData expectedInput
	// unmarshal will call our custom decoders which do input verification
	err := json.Unmarshal([]byte(*incomingJSON), &inputData)
	if err != nil {
		exitLog("program exited due to error in parsing input: " + err.Error())
	}

	config, err := getConfig("")
	if err != nil {
		exitLog("program exited due to error: " + err.Error())
	}

	// lets go
	rawResults, err := config.process(&inputData)
	if err != nil {
		exitLog("program exited due to error: " + err.Error())
	}

	// dump result to STDOUT
	fmt.Println(string(rawResults))

}
