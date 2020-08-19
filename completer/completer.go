/* Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package completer

import (
    "strings"
)

var cmds = map[string][]string {
    "GO": []string{"FROM"},
    "SHOW": []string{"HOSTS", "SPACES", "PARTS", "TAGS", "EDGES", "USERS", "ROLES", "USER", "CONFIGS"},
    "DESCRIBE": []string{"TAG", "EDGE", "SPACE"},
    "DESC": []string{"TAG", "EDGE", "SPACE"},
    "GET": []string{"CONFIGS"},
    "CREATE": []string{"SPACE", "TAG", "EDGE", "USER"},
    "DROP": []string{"SPACE", "TAG", "EDGE", "USER"},
    "ALTER": []string{"TAG", "EDGE", "USER"},
    "INSERT": []string{"VERTEX", "EDGE"},
    "UPDATE": []string{"VERTEX", "EDGE", "CONFIGS"},
    "UPSERT": []string{"VERTEX", "EDGE"},
    "DELETE": []string{"VERTEX", "EDGE"},
    "GRANT": []string{"ROLE"},
    "REVOKE": []string{"ROLE"},
    "CHANGE": []string{"PASSWORD"},

    "FROM": []string{},
    "HOSTS": []string{},
    "SPACES": []string{},
    "PARTS": []string{},
    "TAGS": []string{},
    "EDGES": []string{},
    "USERS": []string{},
    "ROLES": []string{},
    "USER": []string{},
    "CONFIGS": []string{},
    "TAG": []string{},
    "EDGE": []string{},
    "SPACE": []string{},
    "VERTEX": []string{},
    "ROLE": []string{},
    "PASSWORD": []string{},
    "UPTO": []string{},
    "STEPS": []string{},
    "OVER": []string{},
    "AS": []string{},
    "WHERE": []string{},
    "YIELD": []string{},
    "REVERSELY": []string{},
    "VALUES": []string{},
    "OVERWRITE": []string{},
    "IF": []string{},
    "NOT": []string{},
    "EXISTS": []string{},
    "TTL": []string{},
    "PARTITION_NUM": []string{},
    "REPLICA_FACTOR": []string{},
    "TO": []string{},
    "GOD": []string{},
    "ADMIN": []string{},
    "GUEST": []string{},
    "ON": []string{},
    "COUNT": []string{},
    "COUNT_DISTINCT": []string{},
    "SUM": []string{},
    "AVG": []string{},
    "MAX": []string{},
    "MIN": []string{},
    "STD": []string{},
    "BIT_AND": []string{},
    "BIT_OR": []string{},
    "BIT_XOR": []string{},

    // typename
     "int": []string{} ,
     "bool": []string{} ,
     "double": []string{},
     "string": []string{},
     "timestamp": []string{},
     "true": []string{},
     "false": []string{},
}

func NewCompleter(line string, pos int) (string, []string, string) {
    var head = ""
    var completions = []string{}
    var tail = ""
    if len(line) < 1 {
        return head, completions, tail
    }
    words := strings.Fields(line[:pos])
    if len(words) < 1 {
        return head, completions, tail
    }
    cmd := strings.ToUpper(words[len(words)-1])
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
        if subCmds, ok := cmds[cmd]; ok {
            return head, subCmds, tail
        } else {
            return head, completions, tail
        }
    } else {
        for k := range cmds {
            if strings.HasPrefix(k, cmd) {
                completions = append(completions, k)
            }
            //for _, v := range cmds[k] {
            //    if strings.HasPrefix(v, cmd) {
            //        completions = append(completions, v)
            //    }
            //}
        }
        return head, completions, tail
    }
}
