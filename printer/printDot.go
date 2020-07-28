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
			builder.WriteString(fmt.Sprintf("%s->%s[label=\"%t\"];\n", planNodeName, name(condNode), branchInfo.GetIsDoBranch()))
		}
	}
	builder.WriteString("}")
	writer.AppendRow(table.Row{builder.String()})
	fmt.Println(writer.Render())
}
