/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package cli

import (
	"io"
	//"os"

	//"github.com/vesoft-inc/nebula-console/completer"
	"github.com/peterh/liner"
)

var (
    ErrEOF = io.EOF
    ErrAborted = liner.ErrPromptAborted
)

