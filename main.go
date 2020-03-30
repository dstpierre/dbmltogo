package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"unicode"
)

type Database struct {
	XMLName xml.Name `xml:"Database"`
	Tables  []Table  `xml:"Table"`
}

type Table struct {
	XMLName xml.Name `xml:"Table"`
	Name    string   `xml:"Name,attr"`
	Member  string   `xml:"Member,attr"`
	Type    Type     `xml:"Type"`
}

type Type struct {
	XMLName      xml.Name      `xml:"Type"`
	Name         string        `xml:"Name,attr"`
	Columns      []Column      `xml:"Column"`
	Associations []Association `xml:"Association"`
}

type Column struct {
	XMLName       xml.Name `xml:"Column"`
	Name          string   `xml:"Name,attr"`
	CSharpType    string   `xml:"Type,attr"`
	DBType        string   `xml:"DbType,attr"`
	IsPrimaryKey  string   `xml:"IsPrimaryKey,attr"`
	IsDbGenerated string   `xml:"IsDbGenerated,attr"`
	CanBeNull     string   `xml:"CanBeNull,attr"`
}

type Association struct {
	XMLName  xml.Name `xml:"Association"`
	Name     string   `xml:"Name,attr"`
	Member   string   `xml:"Member,attr"`
	ThisKey  string   `xml:"ThisKey,attr"`
	OtherKey string   `xml:"OtherKey,attr"`
	Type     string   `xml:"Type,attr"`
}

var (
	xmlfile        string
	pkgName        string
	elm            bool
	endpointPrefix string
	printTypes     bool
)

func main() {
	flag.StringVar(&xmlfile, "dbml", "", "path to dbml file")
	flag.StringVar(&pkgName, "pkgname", "data", "package name for generated models (default to data)")
	flag.BoolVar(&elm, "elm", false, "generate elm data and endpoint files")
	flag.StringVar(&endpointPrefix, "prefix", "", "Elm API endpoint prefix, i.e. /prefix/membername")
	flag.BoolVar(&printTypes, "printtypes", false, "prints all csharp types from the dbml")
	flag.Parse()

	if len(xmlfile) == 0 || len(pkgName) == 0 {
		flag.Usage()
		return
	}

	if _, err := os.Stat("./" + pkgName); os.IsNotExist(err) {
		if err := os.MkdirAll("./"+pkgName, 0777); err != nil {
			fmt.Printf("cannot create directory ./%s\n", pkgName)
			os.Exit(1)
		}
	}

	if elm {
		if _, err := os.Stat("./elm/Data"); os.IsNotExist(err) {
			if err := os.MkdirAll("./elm/Data", 0777); err != nil {
				fmt.Println("cannot create directory ./elm/Data")
				os.Exit(1)
			}
		}

		if _, err := os.Stat("./elm/Api"); os.IsNotExist(err) {
			if err := os.MkdirAll("./elm/Api", 0777); err != nil {
				fmt.Println("cannot create directory ./elm/Api")
				os.Exit(1)
			}
		}
	}

	b, err := ioutil.ReadFile(xmlfile)
	if err != nil {
		fmt.Println("unable to read dbml file: ", err)
		return
	}

	var db Database
	if err := xml.Unmarshal(b, &db); err != nil {
		fmt.Println("unable to parse dbml file: ", err)
		return
	}

	if printTypes {
		findAllCsharpTypes(db)
		return
	}

	for _, tbl := range db.Tables {
		renderEntity(tbl)

		if elm {
			renderElmEndpoint(tbl, endpointPrefix)
			renderElmData(tbl)
		}
	}

	if err := genNullTypeJSON(); err != nil {
		fmt.Println("error writing custom JSON types file: ", err)
	}

	if err := genScanner(); err != nil {
		fmt.Println("error writing scanner type: ", err)
	}

	if err := renderElmApiEndpoint(); err != nil {
		fmt.Println(err)
	}

	cmd := exec.Command("gofmt", "-w", "./"+pkgName)
	if err := cmd.Run(); err != nil {
		fmt.Println("error running gofmt: ", err)
	}

	cmd = exec.Command("goimports", "-w", "./"+pkgName)
	if err := cmd.Run(); err != nil {
		fmt.Println("error running goimports: ", err)
	}
}

func findAllCsharpTypes(db Database) {
	cstypes := make(map[string]string)
	for _, tbl := range db.Tables {
		for _, c := range tbl.Type.Columns {
			if _, ok := cstypes[c.CSharpType]; !ok {
				cstypes[c.CSharpType] = c.DBType
			}
		}
	}

	for k, v := range cstypes {
		fmt.Println(k, v)
	}
}

func gotype(c Column) (typ string) {
	if c.CanBeNull == "true" {
		switch c.CSharpType {
		case "System.Byte":
			typ = "JSONNullInt32"
		case "System.Int16":
			typ = "JSONNullInt32"
		case "System.Int64":
			typ = "JSONNullInt64"
		case "System.Guid",
			"System.String":
			typ = "JSONNullString"
		case "System.DateTime":
			typ = "JSONNullTime"
		case "System.Single",
			"System.Decimal",
			"System.Double":
			typ = "JSONNullFloat64"
		case "System.Data.Linq.Binary":
			typ = "[]byte"
		case "System.Int32":
			typ = "JSONNullInt32"
		case "System.Boolean":
			typ = "JSONNullBool"
		default:
			fmt.Println("unhandled type: ", c.CSharpType)
		}

		return
	}

	switch c.CSharpType {
	case "System.Byte":
		typ = "byte"
	case "System.Int16":
		typ = "int"
	case "System.Int64":
		typ = "int64"
	case "System.Guid":
		typ = "string"
	case "System.String":
		typ = "string"
	case "System.DateTime":
		typ = "time.Time"
	case "System.Decimal",
		"System.Single",
		"System.Double":
		typ = "float64"
	case "System.Data.Linq.Binary":
		typ = "[]byte"
	case "System.Int32":
		typ = "int"
	case "System.Boolean":
		typ = "bool"
	default:
		fmt.Println("unhandled type: ", c.CSharpType)
	}

	return
}

func camelCase(s string) string {
	s = strings.Replace(s, "ID", "Id", -1)
	return strings.ToLower(s[0:1]) + s[1:]
}

func sanitize(s string) string {
	if strings.HasPrefix(s, "[") {
		s = s[1 : len(s)-2]
		s = strings.Replace(s, " ", "", -1)
	}

	return s
}

func exposedMember(s string) string {
	s = strings.Replace(s, "_", "", -1)

	if unicode.IsDigit(rune(s[0])) {
		s = "DigitStart_" + s
	}

	return strings.ToUpper(s[0:1]) + s[1:]
}
