/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
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

type DataSetPrinter struct {
	writer   table.Writer
	fd       *os.File
	filename string
}

func NewDataSetPrinter() DataSetPrinter {
	writer := table.NewWriter()
	configTableWriter(&writer, false)
	return DataSetPrinter{
		writer: writer,
	}
}

func (p *DataSetPrinter) ExportCsv(filename string) {
	fd, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Printf("Open or Create file %s failed, %s", filename, err.Error())
		return
	}
	p.fd = fd
	p.filename = filename
}

func configWriterTckStyle(writer *table.Writer) {
	(*writer).Style().Format.Header = text.FormatDefault
	(*writer).Style().Options.SeparateRows = false
	(*writer).Style().Options.SeparateHeader = false

	(*writer).Style().Box.MiddleHorizontal = " "
	(*writer).Style().Box.MiddleSeparator = " "
	(*writer).Style().Box.TopLeft = " "
	(*writer).Style().Box.TopRight = " "
	(*writer).Style().Box.TopSeparator = " "
	(*writer).Style().Box.BottomLeft = " "
	(*writer).Style().Box.BottomRight = " "
	(*writer).Style().Box.BottomSeparator = " "
}

func (p *DataSetPrinter) PrintDataSet(res *nebula.ResultSet) {
	if res.GetColSize() == 0 {
		return
	}

	p.writer.ResetHeaders()
	p.writer.ResetRows()

	if res.IsSetPlanDesc() && res.IsSetData() && "tck" == strings.ToLower(string(res.GetPlanDesc().GetFormat())) {
		configWriterTckStyle(&p.writer)
	}

	var header []interface{}
	for _, columName := range res.GetColNames() {
		header = append(header, string(columName))
	}
	p.writer.AppendHeader(table.Row(header))
	numRows := res.GetRowSize()
	numCols := res.GetColSize()
	for i := 0; i < numRows; i++ {
		var newRow []interface{}
		record, err := res.GetRowValuesByIndex(i)
		if err != nil {
			continue
		}
		for j := 0; j < numCols; j++ {
			val, err := record.GetValueByIndex(j)
			if err != nil {
				continue
			}
			newRow = append(newRow, val.String())
		}
		p.writer.AppendRow(table.Row(newRow))
	}

	fmt.Println(p.writer.Render())
	if p.fd != nil {
		s := strings.Replace(p.writer.RenderCSV(), "\\\"", "", -1)
		fmt.Fprintln(p.fd, s)

		if err := p.fd.Close(); err != nil {
			fmt.Printf("Close file %s failed, %s", p.filename, err.Error())
		}
		p.fd = nil
		p.filename = ""
	}

	// Reset the writer style
	if res.IsSetPlanDesc() && res.IsSetData() && "tck" == strings.ToLower(string(res.GetPlanDesc().GetFormat())) {
		p.writer.SetStyle(table.StyleDefault)
		configTableWriter(&p.writer, false)
	}

}
