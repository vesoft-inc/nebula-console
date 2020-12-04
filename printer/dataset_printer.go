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
	fd, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Printf("Open or Create file %s failed, %s", filename, err.Error())
		return
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

func (p *DataSetPrinter) PrintDataSet(res *nebula.ResultSet) {
	if res.IsEmpty() {
		return
	}

	p.writer.ResetHeaders()
	p.writer.ResetRows()
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
		go func() {
			p.fd.Truncate(0)
			p.fd.Seek(0, 0)
			s := strings.Replace(p.writer.RenderCSV(), "\\\"", "", -1)
			fmt.Fprintln(p.fd, s)
		}()
	}
}
