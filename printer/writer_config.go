/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 */

package printer

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func configTableWriter(writer *table.Writer, separateRows bool) {
	(*writer).Style().Format.Header = text.FormatDefault
	(*writer).Style().Options.SeparateRows = separateRows
}
