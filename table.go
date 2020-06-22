/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package main

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	common "github.com/shylock-hg/nebula-go2.0/nebula"
)

func val2String(value *common.Value, depth uint) string {
	// TODO(shylock) get golang runtime limit
	if depth == 0 { // Avoid too deep recursive
		return "..."
	}

	if value.IsSetNVal() { // null
		switch value.GetNVal() {
		case common.NullType___NULL__:
			return "NULL"
		case common.NullType_NaN:
			return "NaN"
		case common.NullType_BAD_DATA:
			return "BAD_DATA"
		case common.NullType_BAD_TYPE:
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
		return "\"" + string(value.GetSVal()) + "\""
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
		// VId only
		return string(value.GetVVal().GetVid())
	} else if value.IsSetEVal() { // Edge
		// src-[TypeName]->dst@ranking
		edge := value.GetEVal()
		return fmt.Sprintf("%s-[%s]->%s@%d", string(edge.GetSrc()), edge.GetName(), string(edge.GetDst()),
			edge.GetRanking())
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
			buffer.WriteString(val2String(v, depth-1))
			buffer.WriteString(",")
		}
		if buffer.Len() > 1 {
			buffer.UnreadRune() // remove last ,
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
			buffer.WriteString(val2String(v, depth-1))
			buffer.WriteString(",")
		}
		if buffer.Len() > 1 {
			buffer.UnreadRune() // remove last ,
		}
		buffer.WriteString("}")
		return buffer.String()
	} else if value.IsSetUVal() { // Set
		// TODO(shylock) optimize the recursive
		s := value.GetUVal()
		var buffer bytes.Buffer
		buffer.WriteString("{")
		for _, v := range s.GetValues() {
			buffer.WriteString(val2String(v, depth-1))
			buffer.WriteString(",")
		}
		if buffer.Len() > 1 {
			buffer.UnreadRune() // remove last ,
		}
		buffer.WriteString("}")
		return buffer.String()
	}
	return ""
}

func max(v1 uint, v2 uint) uint {
	if v1 > v2 {
		return v1
	}
	return v2
}

func sum(a []uint) uint {
	s := uint(0)
	for _, v := range a {
		s += v
	}
	return s
}

type Table struct {
	align        uint   // Each column align indent to boundary
	headerChar   string // Header line characters
	rowChar      string // Row line characters
	colDelimiter string // Column delemiter
}

func NewTable(align uint, header string, row string, delemiter string) Table {
	return Table{align, header, row, delemiter}
}

// Columns width
type TableSpec = []uint
type TableRows = [][]string

func (t Table) printRow(row []string, colSpec TableSpec) {
	for i, col := range row {
		colString := t.colDelimiter + strings.Repeat(" ", int(t.align)) + col
		length := uint(len(col))
		if length < colSpec[i]+t.align {
			colString = colString + strings.Repeat(" ", int(colSpec[i]+t.align-length))
		}
		fmt.Print(colString)
	}
	fmt.Println(t.colDelimiter)
}

func (t Table) PrintTable(table *common.DataSet) {
	columnSize := len(table.GetColumnNames())
	rowSize := len(table.GetRows())
	tableSpec := make(TableSpec, columnSize)
	tableRows := make(TableRows, rowSize)
	tableHeader := make([]string, columnSize)
	for i, header := range table.GetColumnNames() {
		tableSpec[i] = uint(len(header))
		tableHeader[i] = string(header)
	}
	for i, row := range table.GetRows() {
		tableRows[i] = make([]string, columnSize)
		for j, col := range row.GetColumns() {
			tableRows[i][j] = val2String(col, 256)
			tableSpec[j] = max(uint(len(tableRows[i][j])), tableSpec[j])
		}
	}

	//                 value limit         + two indent              + '|' itself
	totalLineLength := int(sum(tableSpec)) + columnSize*int(t.align)*2 + columnSize + 1
	headerLine := strings.Repeat(t.headerChar, totalLineLength)
	rowLine := strings.Repeat(t.rowChar, totalLineLength)
	fmt.Println(headerLine)
	t.printRow(tableHeader, tableSpec)
	fmt.Println(headerLine)
	for _, row := range tableRows {
		t.printRow(row, tableSpec)
		fmt.Println(rowLine)
	}
	fmt.Printf("Got %d rows, %d columns.", rowSize, columnSize)
	fmt.Println()
}
