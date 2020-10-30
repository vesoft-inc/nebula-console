/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package printer

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/vesoft-inc/nebula-go/v2/nebula"
)

type DataSetPrinter struct {
	writer   table.Writer
	fd       *os.File
	filename string
}

func NewDataSetPrinter() DataSetPrinter {
	writer := table.NewWriter()
	configTableWriter(&writer)
	return DataSetPrinter{
		writer: writer,
	}
}

func (p *DataSetPrinter) SetOutCsv(filename string) {
	if p.fd != nil {
		p.UnsetOutCsv()
	}
	fd, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Printf("Open or Create file %s failed, %s", filename, err.Error())
	}
	p.fd = fd
	p.filename = filename
}

func (p *DataSetPrinter) UnsetOutCsv() {
	if p.fd == nil {
		return
	}
	if err := p.fd.Close(); err != nil {
		fmt.Printf("Close file %s failed, %s", p.filename, err.Error())
	}
	p.fd = nil
	p.filename = ""
}

func valueToString(value *nebula.Value) string {
	// TODO(shylock) get golang runtime limit
	if value.IsSetNVal() { // null
		switch value.GetNVal() {
		case nebula.NullType___NULL__:
			return "NULL"
		case nebula.NullType_NaN:
			return "NaN"
		case nebula.NullType_BAD_DATA:
			return "BAD_DATA"
		case nebula.NullType_BAD_TYPE:
			return "BAD_TYPE"
		}
		return "NULL"
	} else if value.IsSetBVal() { // bool
		return strconv.FormatBool(value.GetBVal())
	} else if value.IsSetIVal() { // int64
		return strconv.FormatInt(value.GetIVal(), 10)
	} else if value.IsSetFVal() { // float64
		val := strconv.FormatFloat(value.GetFVal(), 'g', -1, 64)
		if !strings.Contains(val, ".") {
			idx := strings.LastIndex(val, "e")
			if idx == -1 {
				val += ".0"
			} else {
				val = val[0:idx] + ".0" + val[idx:]
			}
		}
		return val
	} else if value.IsSetSVal() { // string
		return `"` + string(value.GetSVal()) + `"`
	} else if value.IsSetDVal() { // yyyy-mm-dd
		date := value.GetDVal()
		str := fmt.Sprintf("%d-%d-%d", date.GetYear(), date.GetMonth(), date.GetDay())
		return str
	} else if value.IsSetTVal() {
		time := value.GetTVal()
		str := fmt.Sprintf("%d:%d:%d:%d",
			time.GetHour(), time.GetMinute(), time.GetSec(), time.GetMicrosec())
		return str
	} else if value.IsSetDtVal() { // yyyy-mm-dd HH:MM:SS:MS TZ
		datetime := value.GetDtVal()
		str := fmt.Sprintf("%d-%d-%d %d:%d:%d:%d",
			datetime.GetYear(), datetime.GetMonth(), datetime.GetDay(),
			datetime.GetHour(), datetime.GetMinute(), datetime.GetSec(), datetime.GetMicrosec())
		return str
	} else if value.IsSetVVal() { // Vertex
		var buffer bytes.Buffer
		vertex := value.GetVVal()
		buffer.WriteString(`("`)
		buffer.WriteString(string(vertex.GetVid()))
		buffer.WriteString(`")`)
		filled := false
		for _, tag := range vertex.GetTags() {
			tagName := string(tag.GetName())
			buffer.WriteString(" :")
			buffer.WriteString(tagName)
			buffer.WriteString("{")
			for k, v := range tag.GetProps() {
				filled = true
				buffer.WriteString(k)
				buffer.WriteString(":")
				buffer.WriteString(valueToString(v))
				buffer.WriteString(", ")
			}
		}
		if filled {
			// remove last ,
			buffer.Truncate(buffer.Len() - 2)
		}
		buffer.WriteString("}")
		return buffer.String()
	} else if value.IsSetEVal() { // Edge
		// (src)-[edge]->(dst)@ranking {props}
		edge := value.GetEVal()
		var buffer bytes.Buffer
		filled := false
		src := string(edge.GetSrc())
		dst := string(edge.GetDst())
		if edge.GetType() < 0 {
			src, dst = dst, src
		}
		buffer.WriteString(fmt.Sprintf(`("%s")-[%s]->("%s")@%d`,
			src, edge.GetName(), dst, edge.GetRanking()))
		buffer.WriteString(" {")
		for k, v := range edge.GetProps() {
			filled = true
			buffer.WriteString(k)
			buffer.WriteString(":")
			buffer.WriteString(valueToString(v))
			buffer.WriteString(", ")
		}
		if filled {
			buffer.Truncate(buffer.Len() - 2)
		}
		buffer.WriteString("}")
		return buffer.String()
	} else if value.IsSetPVal() { // Path
		// src-[TypeName]->dst@ranking-[TypeName]->dst@ranking ...
		p := value.GetPVal()
		str := string(p.GetSrc().GetVid())
		for _, step := range p.GetSteps() {
			pStr := fmt.Sprintf("-[%s]->%s@%d", step.GetName(), string(step.GetDst().GetVid()), step.GetRanking())
			str += pStr
		}
		return str
	} else if value.IsSetLVal() { // List
		// TODO(shylock) optimize the recursive
		l := value.GetLVal()
		var buffer bytes.Buffer
		buffer.WriteString("[")
		for _, v := range l.GetValues() {
			buffer.WriteString(valueToString(v))
			buffer.WriteString(",")
		}
		if buffer.Len() > 1 {
			buffer.Truncate(buffer.Len() - 1)
		}
		buffer.WriteString("]")
		return buffer.String()
	} else if value.IsSetMVal() { // Map
		// TODO(shylock) optimize the recursive
		m := value.GetMVal()
		var buffer bytes.Buffer
		buffer.WriteString("{")
		for k, v := range m.GetKvs() {
			buffer.WriteString("\"" + k + "\"")
			buffer.WriteString(":")
			buffer.WriteString(valueToString(v))
			buffer.WriteString(",")
		}
		if buffer.Len() > 1 {
			buffer.Truncate(buffer.Len() - 1)
		}
		buffer.WriteString("}")
		return buffer.String()
	} else if value.IsSetUVal() { // Set
		// TODO(shylock) optimize the recursive
		s := value.GetUVal()
		var buffer bytes.Buffer
		buffer.WriteString("{")
		for _, v := range s.GetValues() {
			buffer.WriteString(valueToString(v))
			buffer.WriteString(",")
		}
		if buffer.Len() > 1 {
			buffer.Truncate(buffer.Len() - 1)
		}
		buffer.WriteString("}")
		return buffer.String()
	}
	return ""
}

func (p *DataSetPrinter) PrintDataSet(dataset *nebula.DataSet) {
	if len(dataset.GetRows()) == 0 {
		return
	}

	p.writer.ResetHeaders()
	p.writer.ResetRows()
	var header []interface{}
	for _, columName := range dataset.GetColumnNames() {
		header = append(header, string(columName))
	}
	p.writer.AppendHeader(table.Row(header))
	for _, row := range dataset.GetRows() {
		var newRow []interface{}
		for _, column := range row.GetValues() {
			newRow = append(newRow, valueToString(column))
		}
		p.writer.AppendRow(table.Row(newRow))
	}

	fmt.Println(p.writer.Render())
	if p.fd != nil {
		go func() {
			fmt.Fprintln(p.fd, p.writer.RenderCSV())
			fmt.Fprintln(p.fd)
		}()
	}
}
