/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package printer

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/vesoft-inc/nebula-go/nebula/graph"
)

func graphvizString(s string) string {
	s = strings.Replace(s, "{", "\\{", -1)
	s = strings.Replace(s, "}", "\\}", -1)
	s = strings.Replace(s, "\"", "\\\"", -1)
	s = strings.Replace(s, "[", "\\[", -1)
	s = strings.Replace(s, "]", "\\]", -1)
	return s
}

type PlanDescPrinter struct {
	writer   table.Writer
	planDesc *graph.PlanDescription
	fd       *os.File
	filename string
}

func NewPlanDescPrinter() PlanDescPrinter {
	writer := table.NewWriter()
	configTableWriter(&writer)
	return PlanDescPrinter{
		writer:   writer,
		planDesc: nil,
	}
}

func (p *PlanDescPrinter) SetOutDot(filename string) {
	if p.fd != nil {
		p.UnsetOutDot()
	}
	fd, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Printf("Open or Create file %s failed, %s", filename, err.Error())
		return
	}
	p.fd = fd
	p.filename = filename
}

func (p *PlanDescPrinter) UnsetOutDot() {
	if p.fd == nil {
		return
	}
	if err := p.fd.Close(); err != nil {
		fmt.Printf("Close file %s failed, %s", p.filename, err.Error())
	}
	p.fd = nil
	p.filename = ""
}

func name(planNodeDesc *graph.PlanNodeDescription) string {
	return fmt.Sprintf("%s_%d", planNodeDesc.GetName(), planNodeDesc.GetId())
}

func condEdgeLabel(condNode *graph.PlanNodeDescription, doBranch bool) string {
	name := strings.ToLower(string(condNode.GetName()))
	if strings.HasPrefix(name, "select") {
		if doBranch {
			return "Y"
		}
		return "N"
	}
	if strings.HasPrefix(name, "loop") {
		if doBranch {
			return "Do"
		}
	}
	return ""
}

func nodeString(planNodeDesc *graph.PlanNodeDescription, planNodeName string) string {
	var outputVar = graphvizString(string(planNodeDesc.GetOutputVar()))
	var inputVar string
	if planNodeDesc.IsSetDescription() {
		desc := planNodeDesc.GetDescription()
		for _, pair := range desc {
			key := string(pair.GetKey())
			if key == "inputVar" {
				inputVar = graphvizString(string(pair.GetValue()))
			}
		}
	}
	return fmt.Sprintf("\t\"%s\"[label=\"{%s|outputVar: %s|inputVar: %s}\", shape=Mrecord];\n",
		planNodeName, planNodeName, outputVar, inputVar)
}

func edgeString(start, end string) string {
	return fmt.Sprintf("\t\"%s\"->\"%s\";\n", start, end)
}

func conditionalEdgeString(start, end, label string) string {
	return fmt.Sprintf("\t\"%s\"->\"%s\"[label=\"%s\", style=dashed];\n", start, end, label)
}

func conditionalNodeString(name string) string {
	return fmt.Sprintf("\t\"%s\"[shape=diamond];\n", name)
}

func (p PlanDescPrinter) configWriterDotRenderStyle(renderByDot bool) {
	if renderByDot {
		p.writer.Style().Box.Left = " "
		p.writer.Style().Box.Right = " "
	} else {
		p.writer.Style().Box.Left = "|"
		p.writer.Style().Box.Right = "|"
	}
	p.writer.Style().Box.BottomLeft = "-"
	p.writer.Style().Box.BottomRight = "-"
	p.writer.Style().Box.TopLeft = "-"
	p.writer.Style().Box.TopRight = "-"
	p.writer.Style().Box.LeftSeparator = "-"
	p.writer.Style().Box.RightSeparator = "-"
}

func (p PlanDescPrinter) nodeById(nodeId int64) *graph.PlanNodeDescription {
	line := p.planDesc.GetNodeIndexMap()[nodeId]
	return p.planDesc.GetPlanNodeDescs()[line]
}

func (p PlanDescPrinter) makeDotGraphByStruct() string {
	planNodeDescs := p.planDesc.GetPlanNodeDescs()
	var builder strings.Builder
	builder.WriteString("digraph exec_plan {\n")
	builder.WriteString("\trankdir=BT;\n")
	for _, planNodeDesc := range planNodeDescs {
		planNodeName := name(planNodeDesc)
		switch strings.ToLower(string(planNodeDesc.GetName())) {
		case "select":
			builder.WriteString(conditionalNodeString(planNodeName))
		case "loop":
			builder.WriteString(conditionalNodeString(planNodeName))
		default:
			builder.WriteString(nodeString(planNodeDesc, planNodeName))
		}

		if planNodeDesc.IsSetDependencies() {
			for _, depId := range planNodeDesc.GetDependencies() {
				dep := p.nodeById(depId)
				builder.WriteString(edgeString(name(dep), planNodeName))
			}
		}

		if planNodeDesc.IsSetBranchInfo() {
			branchInfo := planNodeDesc.GetBranchInfo()
			condNode := p.nodeById(branchInfo.GetConditionNodeID())
			label := condEdgeLabel(condNode, branchInfo.GetIsDoBranch())
			builder.WriteString(conditionalEdgeString(planNodeName, name(condNode), label))
		}
	}
	builder.WriteString("}")
	return builder.String()
}

func (p PlanDescPrinter) renderDotGraphByStruct(s string) string {
	p.writer.ResetHeaders()
	p.writer.ResetRows()
	p.configWriterDotRenderStyle(true)
	p.writer.AppendHeader(table.Row{"plan"})
	p.writer.AppendRow(table.Row{s})
	return p.writer.Render()
}

func (p PlanDescPrinter) findBranchEndNode(condNodeId int64, isDoBranch bool) int64 {
	for _, node := range p.planDesc.GetPlanNodeDescs() {
		if node.IsSetBranchInfo() {
			bInfo := node.GetBranchInfo()
			if bInfo.GetConditionNodeID() == condNodeId && bInfo.GetIsDoBranch() == isDoBranch {
				return node.GetId()
			}
		}
	}
	return -1
}

func (p PlanDescPrinter) findFirstStartNodeFrom(nodeId int64) int64 {
	node := p.nodeById(nodeId)
	for {
		deps := node.GetDependencies()
		if len(deps) == 0 {
			if strings.ToLower(string(node.GetName())) != "start" {
				return -1
			}
			return node.GetId()
		}
		node = p.nodeById(deps[0])
	}
}

func (p PlanDescPrinter) makeDotGraph() string {
	planNodeDescs := p.planDesc.GetPlanNodeDescs()
	var builder strings.Builder
	builder.WriteString("digraph exec_plan {\n")
	builder.WriteString("\trankdir=BT;\n")
	for _, planNodeDesc := range planNodeDescs {
		planNodeName := name(planNodeDesc)
		switch strings.ToLower(string(planNodeDesc.GetName())) {
		case "select":
			builder.WriteString(conditionalNodeString(planNodeName))
			dep := p.nodeById(planNodeDesc.GetDependencies()[0])
			// then branch
			thenNodeId := p.findBranchEndNode(planNodeDesc.GetId(), true)
			builder.WriteString(edgeString(name(p.nodeById(thenNodeId)), name(dep)))
			thenStartId := p.findFirstStartNodeFrom(thenNodeId)
			builder.WriteString(conditionalEdgeString(name(planNodeDesc), name(p.nodeById(thenStartId)), "Y"))
			// else branch
			elseNodeId := p.findBranchEndNode(planNodeDesc.GetId(), false)
			builder.WriteString(edgeString(name(p.nodeById(elseNodeId)), name(dep)))
			elseStartId := p.findFirstStartNodeFrom(elseNodeId)
			builder.WriteString(conditionalEdgeString(name(planNodeDesc), name(p.nodeById(elseStartId)), "N"))
		case "loop":
			builder.WriteString(conditionalNodeString(planNodeName))
			dep := p.nodeById(planNodeDesc.GetDependencies()[0])
			// do branch
			doNodeId := p.findBranchEndNode(planNodeDesc.GetId(), true)
			builder.WriteString(edgeString(name(p.nodeById(doNodeId)), name(planNodeDesc)))
			doStartId := p.findFirstStartNodeFrom(doNodeId)
			builder.WriteString(conditionalEdgeString(name(planNodeDesc), name(p.nodeById(doStartId)), "Do"))
			// dep
			builder.WriteString(edgeString(name(dep), planNodeName))
		default:
			builder.WriteString(nodeString(planNodeDesc, planNodeName))
			if planNodeDesc.IsSetDependencies() {
				for _, depId := range planNodeDesc.GetDependencies() {
					builder.WriteString(edgeString(name(p.nodeById(depId)), planNodeName))
				}
			}

		}
	}
	builder.WriteString("}")
	return builder.String()
}

func (p PlanDescPrinter) renderDotGraph(s string) string {
	p.writer.ResetHeaders()
	p.writer.ResetRows()
	p.configWriterDotRenderStyle(true)
	p.writer.AppendHeader(table.Row{"plan"})
	p.writer.AppendRow(table.Row{s})
	return p.writer.Render()
}

func (p PlanDescPrinter) renderByRow() string {
	p.writer.ResetHeaders()
	p.writer.ResetRows()
	p.configWriterDotRenderStyle(false)
	planNodeDescs := p.planDesc.GetPlanNodeDescs()

	p.writer.AppendHeader(table.Row{
		"id",
		"name",
		"dependencies",
		"profiling data",
		"operator info",
	})

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

		if planNodeDesc.IsSetProfiles() {
			var strArr []string
			for i, profile := range planNodeDesc.GetProfiles() {
				otherStats := profile.GetOtherStats()
				if otherStats != nil {
					strArr = append(strArr, "{")
				}
				s := fmt.Sprintf("ver: %d, rows: %d, execTime: %dus, totalTime: %dus",
					i, profile.GetRows(), profile.GetExecDurationInUs(), profile.GetTotalDurationInUs())
				strArr = append(strArr, s)

				for k, v := range otherStats {
					strArr = append(strArr, fmt.Sprintf("%s: %s", k, v))
				}
				if otherStats != nil {
					strArr = append(strArr, "}")
				}
			}
			row = append(row, strings.Join(strArr, "\n"))
		} else {
			row = append(row, "")
		}

		var columnInfo []string
		if planNodeDesc.IsSetBranchInfo() {
			branchInfo := planNodeDesc.GetBranchInfo()
			columnInfo = append(columnInfo, fmt.Sprintf("branch: %t, nodeId: %d\n",
				branchInfo.GetIsDoBranch(), branchInfo.GetConditionNodeID()))
		}

		outputVar := fmt.Sprintf("outputVar: %s", string(planNodeDesc.GetOutputVar()))
		columnInfo = append(columnInfo, outputVar)

		if planNodeDesc.IsSetDescription() {
			desc := planNodeDesc.GetDescription()
			for _, pair := range desc {
				columnInfo = append(columnInfo, fmt.Sprintf("%s: %s", string(pair.GetKey()), string(pair.GetValue())))
			}
		}
		row = append(row, strings.Join(columnInfo, "\n"))
		p.writer.AppendRow(table.Row(row))
	}
	return p.writer.Render()
}

func (p *PlanDescPrinter) PrintPlanDesc(planDesc *graph.PlanDescription) {
	p.planDesc = planDesc
	var s string
	format := strings.ToLower(string(planDesc.GetFormat()))
	switch format {
	case "row":
		fmt.Println(p.renderByRow())
	case "dot":
		s = p.makeDotGraph()
		fmt.Println(p.renderDotGraph(s))
	case "dot:struct":
		s = p.makeDotGraphByStruct()
		fmt.Println(p.renderDotGraphByStruct(s))
	}

	outputDot := format != "row"
	if p.fd != nil && outputDot {
		go func() {
			p.fd.Truncate(0)
			p.fd.Seek(0, 0)
			fmt.Fprintln(p.fd, s)
		}()
	}
}
