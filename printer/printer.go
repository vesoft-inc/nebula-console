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

	"github.com/jedib0t/go-pretty/table"
	common "github.com/vesoft-inc/nebula-go/v2/nebula"
	graph "github.com/vesoft-inc/nebula-go/v2/nebula/graph"
)

var writer table.Writer

func init() {
	writer := table.NewWriter()
	writer.SetOutputMirror(os.Stdout)
}

func valueToString(value *common.Value, depth uint) string {
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

func PrintDataSet(dataset *common.DataSet) {
	var header table.Row
	for _, columName := range dataset.GetColumnNames() {
		header = append(header, columName)
	}
	writer.AppendHeader(header)

	for _, row := range dataset.GetRows() {
		var newRow table.Row
		for _, column := range row.GetValues() {
			newRow = append(newRow, valueToString(column, 256))
		}
		writer.AppendRow(newRow)
	}

	writer.Render()
}

func PrintPlanDesc(planDesc *graph.PlanDescription) {
	planNodeDescs := planDesc.GetPlanNodeDescs()

	header := table.Row{"id", "name", "dependencies", "output_var"}
	hasBranchInfo, hasProfilingData, hasDescription := false, false, false
	for _, planNodeDesc := range planNodeDescs {
		var row table.Row
		row = append(row, planNodeDesc.GetId(), planNodeDesc.GetName())

		if planNodeDesc.IsSetDependencies() {
			deps := make([]string, len(planNodeDesc.GetDependencies()))
			for _, dep := range planNodeDesc.GetDependencies() {
				deps = append(deps, fmt.Sprintf("%d", dep))
			}
			row = append(row, strings.Join(deps, ","))
		} else {
			row = append(row, "")
		}

		row = append(row, planNodeDesc.GetOutputVar())

		if planNodeDesc.IsSetBranchInfo() {
			if !hasBranchInfo {
				hasBranchInfo = true
				header = append(header, "branch_info")
			}
			branchInfo := planNodeDesc.GetBranchInfo()
			row = append(row, fmt.Sprintf("do_branch: %b, cond_node_id: %d",
				branchInfo.GetIsDoBranch(), branchInfo.GetConditionNodeID()))
		}

		if planNodeDesc.IsSetProfiles() {
			if !hasProfilingData {
				hasProfilingData = true
				header = append(header, "profiling_data")
			}

			strArr := make([]string, len(planNodeDesc.GetProfiles()))
			for i, profile := range planNodeDesc.GetProfiles() {
				s := fmt.Sprintf("version: %d, num_rows: %d, exec_duration: %dus, total_duration: %dus",
					i, profile.GetRows(), profile.GetExecDurationInUs(), profile.GetTotalDurationInUs())
				strArr = append(strArr, s)
			}
			row = append(row, strings.Join(strArr, ";"))
		}

		if planNodeDesc.IsSetDescription() {
			if !hasDescription {
				hasDescription = true
				header = append(header, "description")
			}
			desc := planNodeDesc.GetDescription()
			var str []string
			for k, v := range desc {
				str = append(str, fmt.Sprintf("%s: %s", k, v))
			}
			row = append(row, strings.Join(str, ","))
		}
	}
	writer.AppendHeader(header)
}
