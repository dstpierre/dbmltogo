package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"
)

var (
	elmDataTemplate     *template.Template
	elmEndpointTemplate *template.Template
)

func init() {
	b, err := ioutil.ReadFile("./data.elm")
	if err != nil {
		log.Fatal(err)
	}

	elmDataTemplate = template.Must(template.New("data").Parse(string(b)))

	b, err = ioutil.ReadFile("./endpoint.elm")
	if err != nil {
		log.Fatal(err)
	}

	elmEndpointTemplate = template.Must(template.New("endpoint").Parse(string(b)))
}

type NameType struct {
	Name     string
	Type     string
	JSONName string
	Nullable bool
}

type ElmData struct {
	ModuleName    string
	APIModuleName string
	PKName        string
	PKType        string
	PKEncodeType  string
	Fields        []NameType
	DecodeFields  []NameType
	EncodeFields  []NameType
}

func renderElmData(tbl Table) {
	pk := getPrimary(tbl.Type.Columns)
	if pk == nil {
		pk = &Column{
			CSharpType:   "System.Int64",
			IsPrimaryKey: "yes",
			Name:         "ID",
		}
	}

	data := ElmData{
		ModuleName:    exposedMember(tbl.Type.Name),
		APIModuleName: exposedMember(tbl.Member),
		PKName:        camelCase(pk.Name),
		PKType:        elmtype(*pk),
		PKEncodeType:  strings.ToLower(elmtype(*pk)),
	}

	for _, c := range tbl.Type.Columns {
		data.DecodeFields = append(data.DecodeFields, NameType{
			Name:     camelCase(c.Name),
			Type:     strings.ToLower(elmtype(c)),
			Nullable: c.CanBeNull == "true",
		})

		if c.Name != pk.Name {
			data.Fields = append(data.Fields, NameType{
				Name:     camelCase(c.Name),
				Type:     elmtype(c),
				Nullable: c.CanBeNull == "true",
			})

			data.EncodeFields = append(data.EncodeFields, NameType{
				Name:     camelCase(c.Name),
				Type:     strings.ToLower(elmtype(c)),
				Nullable: c.CanBeNull == "true",
			})
		}
	}

	var buf bytes.Buffer
	if err := elmDataTemplate.Execute(&buf, data); err != nil {
		fmt.Println("error rendering Elm data template: ", err)
		os.Exit(1)
	}

	filename := fmt.Sprintf("./elm/Data/%s.elm", data.ModuleName)
	if err := ioutil.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		fmt.Println("error writing Elm data module file: ", filename, err)
		os.Exit(1)
	}
}

type ElmEndpoint struct {
	ModuleName string
	Prefix     string
	Endpoint   string
}

func renderElmApiEndpoint() error {
	b, err := ioutil.ReadFile("./api.elm")
	if err != nil {
		return fmt.Errorf("error reading api.elm file: %v", err)
	}

	if err := ioutil.WriteFile("./elm/Api/Endpoint.elm", b, 0644); err != nil {
		return fmt.Errorf("error writing elm/Api/Endpoint.elm: %v", err)
	}
	return nil
}

func renderElmEndpoint(tbl Table, prefix string) {
	data := ElmEndpoint{
		ModuleName: exposedMember(tbl.Member),
		Prefix:     prefix,
		Endpoint:   strings.ToLower(tbl.Member),
	}

	var buf bytes.Buffer
	if err := elmEndpointTemplate.Execute(&buf, data); err != nil {
		fmt.Println("error rendering Elm endpoint template: ", err)
		os.Exit(1)
	}

	filename := fmt.Sprintf("./elm/Api/%s.elm", data.ModuleName)
	if err := ioutil.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		fmt.Println("error writing Elm endpoint file: ", filename, err)
		os.Exit(1)
	}
}

func elmtype(c Column) (typ string) {
	switch c.CSharpType {
	case "System.Byte",
		"System.Int16",
		"System.Int32",
		"System.Int64":
		typ = "Int"
	case "System.Guid",
		"System.String":
		typ = "String"
	case "System.DateTime":
		typ = "Time.Posix"
	case "System.Single",
		"System.Decimal",
		"System.Double":
		typ = "Float"
	case "System.Data.Linq.Binary":
		typ = "Bits"
	case "System.Boolean":
		typ = "Bool"
	default:
		fmt.Println("unhandled type: ", c.CSharpType)
	}

	return

}
