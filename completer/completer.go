/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 */

package completer

import (
	"strings"
)

// all keywords
var keywords = []string{
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
	"CASE", "MATCH", "UNWIND", "SKIP", "SIGN",
	"HOSTS", "SPACE", "SPACES", "VALUES", "USER", "USERS", "PASSWORD",
	"ROLE", "ROLES", "GOD", "ADMIN", "DBA", "GUEST", "GROUP", "PARTITION_NUM",
	"REPLICA_FACTOR", "VID_SIZE", "CHARSET", "COLLATE", "COLLATION", "ALL",
	"LEADER", "UUID", "DATA", "SNAPSHOT", "SNAPSHOTS", "OFFLINE", "ACCOUNT",
	"JOBS", "JOB", "COUNT", "COUNT_DISTINCT", "SUM", "AVG", "MAX", "MIN",
	"STD", "BIT_AND", "BIT_OR", "BIT_XOR", "PATH", "BIDIRECT", "STATUS", "FORCE",
	"PART", "PARTS", "DEFAULT", "HDFS", "CONFIGS", "TTL_DURATION", "TTL_COL",
	"GRAPH", "META", "STORAGE", "SHORTEST", "OUT", "BOTH", "SUBGRAPH", "CONTAINS",
	"TRUE", "FALSE", "THEN", "ELSE", "END", "STARTS", "ENDS", "WITH",
	"<-", "->", "_id", "_type", "_src", "_dst", "_rank", "$$", "$^", "$-",
	/* ".", ",", ":", ";", "@", "+", "-", "*", "/", "%", "!", "^", "<", "<=",
	   ">", ">=", "==", "!=", "||", "&&", "|", "=", "(", ")", "[", "]" */
}

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

	/* BALANCE */
	"BALANCE": []string{"LEADER", "DATA", "STOP"},

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
	"STARTS":    []string{"WITH"},
	"ENDS":      []string{"WITH"},

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
	if len(line) < 1 {
		return
	}
	words := strings.Fields(line[:pos])
	if len(words) < 1 {
		return
	}
	lastWord := strings.ToUpper(words[len(words)-1])
	h := strings.LastIndex(line[:pos], " ")
	head = line[:h+1]
	tail = line[pos:]
	if line[pos-1] == ' ' { // find sub cmd
		if subs, ok := subCmds[lastWord]; ok {
			completions = append(completions, subs...)
		}
	} else {
		for _, k := range keywords {
			if strings.HasPrefix(k, lastWord) {
				completions = append(completions, k)
			}
		}
	}

	if len(completions) == 1 {
		completions[0] += " "
	}

	return
}
