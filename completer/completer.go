/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package completer

import "github.com/vesoft-inc/readline"

var prefixCompleter = readline.NewPrefixCompleter(
	// show
	readline.PcItem("SHOW",
		readline.PcItem("HOSTS"),
		readline.PcItem("SPACES"),
		readline.PcItem("PARTS"),
		readline.PcItem("TAGS"),
		readline.PcItem("EDGES"),
		readline.PcItem("USERS"),
		readline.PcItem("ROLES"),
		readline.PcItem("USER"),
		readline.PcItem("CONFIGS"),
	),

	// describe
	readline.PcItem("DESCRIBE",
		readline.PcItem("TAG"),
		readline.PcItem("EDGE"),
		readline.PcItem("SPACE"),
	),
	readline.PcItem("DESC",
		readline.PcItem("TAG"),
		readline.PcItem("EDGE"),
		readline.PcItem("SPACE"),
	),
	// get configs
	readline.PcItem("GET",
		readline.PcItem("CONFIGS"),
	),
	// create
	readline.PcItem("CREATE",
		readline.PcItem("SPACE"),
		readline.PcItem("TAG"),
		readline.PcItem("EDGE"),
		readline.PcItem("USER"),
	),
	// drop
	readline.PcItem("DROP",
		readline.PcItem("SPACE"),
		readline.PcItem("TAG"),
		readline.PcItem("EDGE"),
		readline.PcItem("USER"),
	),
	// alter
	readline.PcItem("ALTER",
		readline.PcItem("USER"),
		readline.PcItem("TAG"),
		readline.PcItem("EDGE"),
	),

	// insert
	readline.PcItem("INSERT",
		readline.PcItem("VERTEX"),
		readline.PcItem("EDGE"),
	),
	// update
	readline.PcItem("UPDATE",
		readline.PcItem("CONFIGS"),
		readline.PcItem("VERTEX"),
		readline.PcItem("EDGE"),
	),
	// upsert
	readline.PcItem("UPSERT",
		readline.PcItem("VERTEX"),
		readline.PcItem("EDGE"),
	),
	// delete
	readline.PcItem("DELETE",
		readline.PcItem("VERTEX"),
		readline.PcItem("EDGE"),
	),

	// grant
	readline.PcItem("GRANT",
		readline.PcItem("ROLE"),
	),
	// revoke
	readline.PcItem("REVOKE",
		readline.PcItem("ROLE"),
	),
	// change password
	readline.PcItem("CHANGE",
		readline.PcItem("PASSWORD"),
	),
)

func NewCompleter() *readline.PrefixCompleter {
	return prefixCompleter
}
