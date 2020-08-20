/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package completer

import (
	"strings"
)

// all keywords
var cmds = []string{
	/* Reserved keyword */
	"GO", "AS", "TO", "OR", "AND", "XOR", "USE", "SET", "FROM",
	"WHERE", "MATCH", "INSERT", "YIELD", "RETURN", "DESCRIBE",
	"DESC", "VERTEX", "EDGE", "EDGES", "UPDATE", "UPSERT",
	"WHEN", "DELETE", "FIND", "LOOKUP", "ALTER", "STEPS", "OVER",
	"UPTO", "REVERSELY", "INDEX", "INDEXES", "REBUILD", "BOOL",
	"INT8", "INT16", "INT32", "INT64", "INT", "FLOAT", "DOUBLE",
	"STRING", "FIXED_STRING", "TIMESTAMP", "DATE", "DATETIME",
	"TAG", "TAGS", "UNION", "INTERSECT", "MINUS", "NO", "OVERWRITE",
	"SHOW", "ADD", "CREATE", "DROP", "REMOVE", "IF", "NOT", "EXISTS",
	"WITH", "CHANGE", "GRANT", "REVOKE", "ON", "BY", "IN", "DOWNLOAD",
	"GET", "OF", "ORDER", "INGEST", "COMPACT", "FLUSH", "SUBMIT",
	"ASC", "DISTINCT", "FETCH", "PROP", "BALANCE", "STOP", "LIMIT",
	"OFFSET", "IS", "NULL", "RECOVER", "EXPLAIN", "PROFILE", "FORMAT",
	/* Unreserved keyword */
	"HOSTS", "SPACE", "SPACES", "VALUES", "USER", "USERS", "PASSWORD",
	"ROLE", "ROLES", "GOD", "ADMIN", "DBA", "GUEST", "GROUP", "PARTITION_NUM",
	"REPLICA_FACTOR", "VID_SIZE", "CHARSET", "COLLATE", "COLLATION", "ALL",
	"LEADER", "UUID", "DATA", "SNAPSHOT", "SNAPSHOTS", "OFFLINE", "ACCOUNT",
	"JOBS", "JOB", "COUNT", "COUNT_DISTINCT", "SUM", "AVG", "MAX", "MIN",
	"STD", "BIT_AND", "BIT_OR", "BIT_XOR", "PATH", "BIDIRECT", "STATUS", "FORCE",
	"PART", "PARTS", "DEFAULT", "HDFS", "CONFIGS", "TTL_DURATION", "TTL_COL",
	"GRAPH", "META", "STORAGE", "SHORTEST", "OUT", "BOTH", "SUBGRAPH", "CONTAINS",
	"TRUE", "FALSE",
	"<-", "->", "_id", "_type", "_src", "_dst", "_rank", "$$", "$^", "$-",
	/* ".", ",", ":", ";", "@", "+", "-", "*", "/", "%", "!", "^", "<", "<=",
	   ">", ">=", "==", "!=", "||", "&&", "|", "=", "(", ")", "[", "]" */
}

// lowercase letters of data types
/*
var types = []string{
    "int", "int8", "int16", "int32", "int64",
    "float", "double",
    "string", "fixed_string",
    "timestamp", "date", "datetime",
    "bool", "true", "false",
}
*/

var subCmds = map[string][]string{
	/* SHOW */
	"SHOW": []string{"TAGS", "EDGES", "SPACES", "HOSTS", "CHARSET", "COLLATION", "USERS",
		"CONFIGS", "CREATE", "INDEXES", "PARTS", "ROLES", "SNAPSHOTS",
		"TAG", "EDGE"},
	"CONFIGS": []string{"GRAPH", "STORAGE", "META"},

	/* FIND PATH */
	"FIND":     []string{"SHORTEST", "ALL"},
	"SHORTEST": []string{"PATH"},
	"ALL":      []string{"PATH"},
	"PATH":     []string{"FROM"},

	/* GROUP BY */
	"GROUP": []string{"BY"},

	/* ORDER BY */
	"ORDER": []string{"BY"},

	/* UNION */
	"UNION": []string{"DISTINCT", "ALL"},

	/* DESCRIBE */
	"DESCRIBE": []string{"TAG", "EDGE", "SPACE"},
	"DESC":     []string{"TAG", "EDGE", "SPACE"},

	/* DDL */
	"ALTER":   []string{"TAG", "EDGE", "USER"},
	"CREATE":  []string{"SPACE", "TAG", "EDGE", "USER"},
	"SPACE":   []string{"IF NOT EXISTS", "IF EXISTS"},
	"TAG":     []string{"INDEX", "IF NOT EXISTS", "IF EXISTS"},
	"EDGE":    []string{"INDEX", "IF NOT EXISTS", "IF EXISTS"},
	"USER":    []string{"IF NOT EXISTS", "IF EXISTS"},
	"INDEX":   []string{"IF NOT EXISTS", "IF EXISTS"},
	"DROP":    []string{"SPACE", "TAG", "EDGE", "USER"},
	"REBUILD": []string{"TAG", "EDGE"},

	/* DQL */
	"GO":        []string{"FROM"},
	"STEPS":     []string{"FROM"},
	"REVERSELY": []string{"BIDIRECT"},
	"BIDIRECT":  []string{"WHERE"},
	"YIELD":     []string{"DISTINCT"},
	"FETCH":     []string{"PROP"},
	"PROP":      []string{"ON"},
	"LOOKUP":    []string{"ON"},

	/* DML */
	"INSERT": []string{"VERTEX", "EDGE"},
	"DELETE": []string{"VERTEX", "EDGE"},
	"UPDATE": []string{"VERTEX", "EDGE"},
	"UPSERT": []string{"VERTEX", "EDGE"},

	/* something about user and role */
	"WITH":   []string{"PASSWORD"},
	"CHANGE": []string{"PASSWORD"},
	"GRANT":  []string{"ROLE"},
	"REVOKE": []string{"ROLE"},
}

func NewCompleter(line string, pos int) (head string, completions []string, tail string) {
	head = ""
	completions = []string{}
	tail = ""
	if len(line) < 1 {
		return head, completions, tail
	}
	words := strings.Fields(line[:pos])
	if len(words) < 1 {
		return head, completions, tail
	}
	find := words[len(words)-1]
	upperFind := strings.ToUpper(find)
	h := strings.LastIndex(line[:pos], " ")
	head = line[:h+1]
	tail = line[pos:]
	var findSub bool
	if line[pos-1] == ' ' {
		findSub = true
	} else {
		findSub = false
	}
	if findSub {
		if subs, ok := subCmds[upperFind]; ok {
			completions = subs
			//return head, subs, tail
		} else {
			//return head, completions, tail
		}
	} else {
		//for _, t := range types {
		//    if strings.HasPrefix(t, find) {
		//        completions = append(completions, t)
		//    }
		//}
		//if len(completions) > 0 {
		//    return head, completions, tail
		//}
		for _, k := range cmds {
			if strings.HasPrefix(k, upperFind) {
				completions = append(completions, k)
			}
			//for _, v := range cmds[k] {
			//    if strings.HasPrefix(v, cmd) {
			//        completions = append(completions, v)
			//    }
			//}
		}
		//return head, completions, tail
	}
	if len(completions) == 1 {
		completions[0] += " "
	}
	return
}
