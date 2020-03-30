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
	goEntityTemplate *template.Template
)

func init() {
	b, err := ioutil.ReadFile("./goentity.txt")
	if err != nil {
		log.Fatal(err)
	}

	goEntityTemplate = template.Must(template.New("goentity").Parse(string(b)))
}

type Entity struct {
	PkgName string
	EntityName string
	MemberName string
	TableName string
	PKName string
	PKType string
	Fields []NameType
	InsertFields []NameType
}

func renderEntity(tbl Table) {
	pk := getPrimary(tbl.Type.Columns)
	if pk == nil {
		pk = &Column{
			CSharpType: "System.Int64",
			IsPrimaryKey: "yes",
			Name: "ID",
		}
	}

	data := Entity{
		PkgName: pkgName,
		EntityName: exposedMember(tbl.Type.Name),
		MemberName: exposedMember(tbl.Member),
		TableName: tbl.Name,
		PKName: pk.Name,
		PKType: gotype(*pk),
	}	

	for _, c := range tbl.Type.Columns {
		data.Fields = append(data.Fields, NameType{
			Name: exposedMember(c.Name),
			Type: gotype(c),
			JSONName: camelCase(c.Name),
		})

		if c.Name == pk.Name {
			continue
		}

		data.InsertFields = append(data.InsertFields, NameType{
			Name: exposedMember(c.Name),
			Type: gotype(c),
		})
	}

	var buf bytes.Buffer
	if err := goEntityTemplate.Execute(&buf, data); err != nil {
		fmt.Println("error rendering Go entity: ", tbl.Name, err)
		os.Exit(1)
	}

	filename := fmt.Sprintf("./%s/%s.go", pkgName, strings.ToLower(tbl.Member))
	if err := ioutil.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		fmt.Println("error writing Go entity file: ", filename, err)
		os.Exit(1)
	}
}

func getPrimary(cols []Column) (pk *Column) {
	for _, col := range cols {
		if col.IsPrimaryKey == "true" {
			pk = &col
			return
		}
	}
	return
}

func genSelectFields(cols []Column) (src string) {
	for _, col := range cols {
		src += fmt.Sprintf("\n\t%s,", col.Name)
	}

	return strings.TrimRight(src, ",")
}

func genNullTypeJSON() error {
	src := fmt.Sprintf(`
package %s

import (
	"encoding/json"
	"database/sql"
	"time"
)

type JSONNullBool struct {
	sql.NullBool
}

func (v JSONNullBool) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Bool)
	} else {
		return json.Marshal(nil)
	}
}

func (v *JSONNullBool) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *bool
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.Bool = *x
	} else {
		v.Valid = false
	}
	return nil
}

type JSONNullFloat64 struct {
	sql.NullFloat64
}

func (v JSONNullFloat64) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Float64)
	} else {
		return json.Marshal(nil)
	}
}

func (v *JSONNullFloat64) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *float64
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.Float64 = *x
	} else {
		v.Valid = false
	}
	return nil
}

type JSONNullInt32 struct {
	sql.NullInt32
}

func (v JSONNullInt32) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Int32)
	} else {
		return json.Marshal(nil)
	}
}

func (v *JSONNullInt32) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *int32
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.Int32 = *x
	} else {
		v.Valid = false
	}
	return nil
}

type JSONNullInt64 struct {
	sql.NullInt64
}

func (v JSONNullInt64) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Int64)
	} else {
		return json.Marshal(nil)
	}
}

func (v *JSONNullInt64) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *int64
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.Int64 = *x
	} else {
		v.Valid = false
	}
	return nil
}

type JSONNullString struct {
	sql.NullString
}

func (v JSONNullString) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.String)
	} else {
		return json.Marshal(nil)
	}
}

func (v *JSONNullString) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *string
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.String = *x
	} else {
		v.Valid = false
	}
	return nil
}

type JSONNullTime struct {
	sql.NullTime
}

func (v JSONNullTime) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Time)
	} else {
		return json.Marshal(nil)
	}
}

func (v *JSONNullTime) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *time.Time
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.Time = *x
	} else {
		v.Valid = false
	}
	return nil
}
	`, pkgName)

	filename := fmt.Sprintf("./%s/types.go", pkgName)
	return ioutil.WriteFile(filename, []byte(src), 0644)

}

func genScanner() error {
	src := fmt.Sprintf(`
package %s

type scanner interface {
	Scan(dest ...interface{}) error
}
	`, pkgName)

	filename := fmt.Sprintf("./%s/scanner.go", pkgName)
	return ioutil.WriteFile(filename, []byte(src), 0644)
}
