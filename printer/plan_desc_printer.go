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
	nebula "github.com/vesoft-inc/nebula-go"
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
	fd       *os.File
	filename string
}

func NewPlanDescPrinter() PlanDescPrinter {
	writer := table.NewWriter()
	configTableWriter(&writer)
	return PlanDescPrinter{
		writer: writer,
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

func (p PlanDescPrinter) renderDotGraph(s string) string {
	p.writer.ResetHeaders()
	p.writer.ResetRows()
	p.configWriterDotRenderStyle(true)
	p.writer.AppendHeader(table.Row{"plan"})
	p.writer.AppendRow(table.Row{s})
	return p.writer.Render()
}

func (p PlanDescPrinter) renderDotGraphByStruct(s string) string {
	p.writer.ResetHeaders()
	p.writer.ResetRows()
	p.configWriterDotRenderStyle(true)
	p.writer.AppendHeader(table.Row{"plan"})
	p.writer.AppendRow(table.Row{s})
	return p.writer.Render()
}

func (p PlanDescPrinter) renderByRow(rows [][]interface{}) string {
	p.writer.ResetHeaders()
	p.writer.ResetRows()
	p.configWriterDotRenderStyle(false)
	p.writer.AppendHeader(table.Row{
		"id",
		"name",
		"dependencies",
		"profiling data",
		"operator info",
	})

	for _, row := range rows {
		p.writer.AppendRow(table.Row(row))
	}
	return p.writer.Render()
}

func (p *PlanDescPrinter) PrintPlanDesc(res *nebula.ResultSet) {
	var s string
	format := strings.ToLower(string(res.GetPlanDesc().GetFormat()))
	switch format {
	case "row":
		rows := res.MakePlanByRow()
		fmt.Println(p.renderByRow(rows))
	case "dot":
		s = res.MakeDotGraph()
		fmt.Println(p.renderDotGraph(s))
	case "dot:struct":
		s = res.MakeDotGraphByStruct()
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
