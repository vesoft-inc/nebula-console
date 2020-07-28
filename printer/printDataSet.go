/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package printer

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/vesoft-inc/nebula-go/v2/nebula"
)

func valueToString(value *nebula.Value, depth uint) string {
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
		return strconv.FormatFloat(value.GetFVal(), 'g', -1, 64)
	} else if value.IsSetSVal() { // string
		return string(value.GetSVal())
	} else if value.IsSetDVal() { // yyyy-mm-dd
		date := value.GetDVal()
		str := fmt.Sprintf("%d-%d-%d", date.GetYear(), date.GetMonth(), date.GetDay())
		return str
	} else if value.IsSetTVal() { // yyyy-mm-dd HH:MM:SS:MS TZ
		datetime := value.GetTVal()
		str := fmt.Sprintf("%d-%d-%d %d:%d:%d:%d UTC%d",
			datetime.GetYear(), datetime.GetMonth(), datetime.GetDay(),
			datetime.GetHour(), datetime.GetMinute(), datetime.GetSec(), datetime.GetMicrosec(),
			datetime.GetTimezone())
		return str
	} else if value.IsSetVVal() { // Vertex
		var buffer bytes.Buffer
		vertex := value.GetVVal()
		buffer.WriteString("(")
		buffer.WriteString(string(vertex.GetVid()))
		buffer.WriteString(")")
		buffer.WriteString(" ")
		filled := false
		for _, tag := range vertex.GetTags() {
			tagName := string(tag.GetName())
			for k, v := range tag.GetProps() {
				filled = true
				buffer.WriteString(tagName)
				buffer.WriteString(".")
				buffer.WriteString(k)
				buffer.WriteString(":")
				buffer.WriteString(valueToString(v, depth-1))
				buffer.WriteString(",")
			}
		}
		if filled {
			// remove last ,
			buffer.Truncate(buffer.Len() - 1)
		}
		return buffer.String()
	} else if value.IsSetEVal() { // Edge
		// src-[TypeName]->dst@ranking
		edge := value.GetEVal()
		var buffer bytes.Buffer
		filled := false
		buffer.WriteString(fmt.Sprintf("%s-[%s]->%s@%d", string(edge.GetSrc()), edge.GetName(), string(edge.GetDst()),
			edge.GetRanking()))
		buffer.WriteString(" ")
		for k, v := range edge.GetProps() {
			filled = true
			buffer.WriteString(k)
			buffer.WriteString(":")
			buffer.WriteString(valueToString(v, depth-1))
			buffer.WriteString(",")
		}
		if filled {
			buffer.Truncate(buffer.Len() - 1)
		}
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
			buffer.WriteString(valueToString(v, depth-1))
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
			buffer.WriteString(valueToString(v, depth-1))
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
			buffer.WriteString(valueToString(v, depth-1))
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

func printDataSet(dataset *nebula.DataSet) {
	writer := table.NewWriter()
	writer.Style().Options.SeparateRows = true

	var header []interface{}
	for _, columName := range dataset.GetColumnNames() {
		header = append(header, string(columName))
	}
	writer.AppendHeader(table.Row(header))

	for _, row := range dataset.GetRows() {
		var newRow []interface{}
		for _, column := range row.GetValues() {
			newRow = append(newRow, valueToString(column, 256))
		}
		writer.AppendRow(table.Row(newRow))
	}

	fmt.Println(writer.Render())
}
