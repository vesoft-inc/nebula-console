/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package printer

import (
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/vesoft-inc/nebula-go/v2/nebula"
	"github.com/vesoft-inc/nebula-go/v2/nebula/graph"
)

func PrintDataSet(dataset *nebula.DataSet) {
	writer := table.NewWriter()

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
		writer.AppendSeparator()
	}

	fmt.Println(writer.Render())
}

func PrintPlanDesc(planDesc *graph.PlanDescription) {
	fmt.Println("\n\nExecution Plan\n")

	switch planDesc.GetFormat() {
	case graph.PlanFormat_ROW:
		printPlanDescByRow(planDesc)
	case graph.PlanFormat_DOT:
		printPlanDescByDot(planDesc)
	}
}
