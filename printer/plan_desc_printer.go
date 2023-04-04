/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.qls
 */

package printer

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	nebula "github.com/vesoft-inc/nebula-go/v3"
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
	configTableWriter(&writer, true)
	return PlanDescPrinter{
		writer: writer,
	}
}

func (p *PlanDescPrinter) ExportExecutionPlan(filename string) {
	fd, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Printf("Open or Create file %s failed, %s", filename, err.Error())
		return
	}
	p.fd = fd
	p.filename = filename
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

func (p PlanDescPrinter) configWriterTckRenderStyle() {
	p.writer.Style().Format.Header = text.FormatDefault
	p.writer.Style().Options.SeparateRows = false
	p.writer.Style().Options.SeparateHeader = false

	p.writer.Style().Box.MiddleHorizontal = " "
	p.writer.Style().Box.MiddleSeparator = " "
	p.writer.Style().Box.TopLeft = " "
	p.writer.Style().Box.TopRight = " "
	p.writer.Style().Box.TopSeparator = " "
	p.writer.Style().Box.BottomLeft = " "
	p.writer.Style().Box.BottomRight = " "
	p.writer.Style().Box.BottomSeparator = " "
}

func (p PlanDescPrinter) renderByTck(rows [][]interface{}) string {
	p.writer.ResetHeaders()
	p.writer.ResetRows()
	p.configWriterTckRenderStyle()
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
		s = p.renderByRow(rows)
		fmt.Println(s)
	case "dot":
		s = res.MakeDotGraph()
		fmt.Println(p.renderDotGraph(s))
	case "dot:struct":
		s = res.MakeDotGraphByStruct()
		fmt.Println(p.renderDotGraphByStruct(s))
	case "tck":
		rows := res.MakePlanByTck()
		fmt.Println(p.renderByTck(rows))
		// Reset the writer style
		p.writer.SetStyle(table.StyleDefault)
		configTableWriter(&p.writer, true)
	}

	if p.fd != nil {
		go func() {
			fmt.Fprintln(p.fd, s)

			if err := p.fd.Close(); err != nil {
				fmt.Printf("Close file %s failed, %s", p.filename, err.Error())
			}
			p.fd = nil
			p.filename = ""
		}()
	}
}
