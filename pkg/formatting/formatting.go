// Copyright © 2019 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package formatting

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"text/template"

	"github.com/olekukonko/tablewriter"
)

type Column struct {
	Name      string
	Template  *template.Template
	MaxLength int
}

type Table struct {
	Columns   []Column
	Rows      []interface{}
	Separator string

	t   *tablewriter.Table
	buf *bytes.Buffer
}

const ellipsis = "…"

func (c *Column) FormatFieldOrError(data interface{}) (string, error) {
	buf := new(bytes.Buffer)
	tpl := c.Template
	err := tpl.Execute(buf, data)

	return buf.String(), err
}

func (c *Column) FormatField(data interface{}) string {
	result, err := c.FormatFieldOrError(data)
	if err != nil {
		return fmt.Sprintf("#(%v)", err)
	}

	return trunc(result, c.MaxLength)
}

func trunc(s string, length int) string {
	if length > 0 && len(s) > length {
		return s[0:length-len(ellipsis)] + ellipsis
	}

	return s
}

func NewColumn(name string) *Column {
	return NamedColumn(name, name)
}

func NamedColumn(name, fieldName string) *Column {
	tpl := fmt.Sprintf("{{.%s}}", fieldName)
	col, err := CustomColumn(name, tpl)
	if err != nil {
		panic(err)
	}

	return col
}

func CustomColumn(name, tpl string) (*Column, error) {
	parsedTemplate, err := template.New(name).Parse(tpl)
	if err != nil {
		return nil, err
	}

	return &Column{Name: name, Template: parsedTemplate}, nil
}

func NewTable(data interface{}, fields []string, headers []string) *Table {
	buf := new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetHeader(headers)

	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.SetAutoFormatHeaders(false)
	table.SetColWidth(60)

	columns := make([]Column, 0, len(fields))
	for i, field := range fields {
		columns = append(columns, *NamedColumn(headers[i], field))
	}

	slice := asSlice(data)

	return &Table{Columns: columns, Rows: slice, Separator: "  ", t: table, buf: buf}
}

func asSlice(slice interface{}) []interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		slice = []interface{}{s}
		s = reflect.ValueOf(slice)
	}

	ret := make([]interface{}, s.Len())
	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret
}

func (t *Table) Format(color bool) string {
	formattedFields := make([][]string, len(t.Rows))
	for i, row := range t.Rows {
		formattedRow := make([]string, len(t.Columns))
		for i, column := range t.Columns {
			value := column.FormatField(row)
			formattedRow[i] = value
		}

		formattedFields[i] = formattedRow
	}

	for _, fields := range formattedFields {
		t.t.Append(fields)
	}
	t.t.Render()

	p := make([]string, 0)
	for _, v := range strings.Split(t.buf.String(), "\n") {
		if len(v) > 1 && v[0:2] == "  " {
			v = v[2:]
		}
		if v != "" {
			p = append(p, v)
		}
	}

	return strings.Join(p, "\n")
}
