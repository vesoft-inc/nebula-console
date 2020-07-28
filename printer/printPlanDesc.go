/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package printer

import (
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/vesoft-inc/nebula-go/v2/nebula/graph"
)

func name(planNodeDesc *graph.PlanNodeDescription) string {
	return fmt.Sprintf("%s_%d", planNodeDesc.GetName(), planNodeDesc.GetId())
}

func condEdgeLabel(condNode *graph.PlanNodeDescription, doBranch bool) string {
	name := strings.ToLower(string(condNode.GetName()))
	if strings.HasPrefix(name, "select") {
		if doBranch {
			return "Yes"
		}
		return "No"
	}
	if strings.HasPrefix(name, "loop") {
		if doBranch {
			return "LoopBody"
		}
	}
	return ""
}

func printPlanDescByDot(planDesc *graph.PlanDescription) {
	writer := table.NewWriter()
	writer.Style().Box.Left = " "
	writer.Style().Box.Right = " "
	writer.Style().Box.BottomLeft = "-"
	writer.Style().Box.BottomRight = "-"
	writer.Style().Box.TopLeft = "-"
	writer.Style().Box.TopRight = "-"
	writer.Style().Box.LeftSeparator = "-"
	writer.Style().Box.RightSeparator = "-"

	writer.AppendHeader(table.Row{"plan"})
	nodeIdxMap := planDesc.GetNodeIndexMap()
	planNodeDescs := planDesc.GetPlanNodeDescs()
	var builder strings.Builder
	builder.WriteString("digraph exec_plan {\n")
	for _, planNodeDesc := range planNodeDescs {
		planNodeName := name(planNodeDesc)
		switch strings.ToLower(string(planNodeDesc.GetName())) {
		case "select":
			builder.WriteString(fmt.Sprintf("%s[shape=diamond];\n", planNodeName))
		case "loop":
			builder.WriteString(fmt.Sprintf("%s[shape=diamond];\n", planNodeName))
		default:
			builder.WriteString(fmt.Sprintf("%s[shape=box, style=rounded];\n", planNodeName))
		}

		if planNodeDesc.IsSetDependencies() {
			for _, depId := range planNodeDesc.GetDependencies() {
				dep := planNodeDescs[nodeIdxMap[depId]]
				builder.WriteString(fmt.Sprintf("%s->%s;\n", name(dep), planNodeName))
			}
		}
		if planNodeDesc.IsSetBranchInfo() {
			branchInfo := planNodeDesc.GetBranchInfo()
			condNode := planNodeDescs[nodeIdxMap[branchInfo.GetConditionNodeID()]]
			builder.WriteString(fmt.Sprintf("%s->%s[label=\"%s\"];\n",
				planNodeName, name(condNode), condEdgeLabel(condNode, branchInfo.GetIsDoBranch())))
		}
	}
	builder.WriteString("}")
	writer.AppendRow(table.Row{builder.String()})
	fmt.Println(writer.Render())
}

func printPlanDescByRow(planDesc *graph.PlanDescription) {
	writer := table.NewWriter()
	writer.Style().Options.SeparateRows = true

	planNodeDescs := planDesc.GetPlanNodeDescs()

	header := table.Row{"id", "name", "dependencies", "output_var", "branch_info", "profiling_data", "description"}
	var rows []table.Row
	for _, planNodeDesc := range planNodeDescs {
		var row []interface{}
		row = append(row, planNodeDesc.GetId(), string(planNodeDesc.GetName()))

		if planNodeDesc.IsSetDependencies() {
			var deps []string
			for _, dep := range planNodeDesc.GetDependencies() {
				deps = append(deps, fmt.Sprintf("%d", dep))
			}
			row = append(row, strings.Join(deps, ","))
		} else {
			row = append(row, "")
		}

		row = append(row, string(planNodeDesc.GetOutputVar()))

		if planNodeDesc.IsSetBranchInfo() {
			branchInfo := planNodeDesc.GetBranchInfo()
			row = append(row, fmt.Sprintf("branch: %t, node_id: %d",
				branchInfo.GetIsDoBranch(), branchInfo.GetConditionNodeID()))
		} else {
			row = append(row, "")
		}

		if planNodeDesc.IsSetProfiles() {
			var strArr []string
			for i, profile := range planNodeDesc.GetProfiles() {
				s := fmt.Sprintf("version: %d, rows: %d, exec_time: %dus, total_time: %dus",
					i, profile.GetRows(), profile.GetExecDurationInUs(), profile.GetTotalDurationInUs())
				strArr = append(strArr, s)
			}
			row = append(row, strings.Join(strArr, "\n"))
		} else {
			row = append(row, "")
		}

		if planNodeDesc.IsSetDescription() {
			desc := planNodeDesc.GetDescription()
			var str []string
			for k, v := range desc {
				str = append(str, fmt.Sprintf("%s: %s", k, string(v)))
			}
			row = append(row, strings.Join(str, ","))
		} else {
			row = append(row, "")
		}
		rows = append(rows, table.Row(row))
	}
	writer.AppendHeader(header)
	writer.AppendRows(rows)
	fmt.Println(writer.Render())
}
