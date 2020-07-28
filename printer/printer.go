/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package printer

import (
	"fmt"

	"github.com/vesoft-inc/nebula-go/v2/nebula"
	"github.com/vesoft-inc/nebula-go/v2/nebula/graph"
)

func PrintDataSet(dataset *nebula.DataSet) {
	printDataSet(dataset)
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
