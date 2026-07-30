package common

import (
	"regexp"
	"strings"
)

// aggFuncRe 匹配聚合函数调用（COUNT/SUM/AVG/MIN/MAX），用于识别未取别名的聚合列。
// application 层预览校验与 job 层采集校验共用同一判据，集中定义避免重复（D-025）。
var aggFuncRe = regexp.MustCompile(`(?i)^(count|sum|avg|min|max)\s*\(`)

// FindUnnamedAggColumns 找出无别名聚合函数列或空列名。
// 列名若为聚合函数调用（如 "COUNT(*)"、"sum(amount)"）或以空串呈现，提示需用 AS 取别名。
func FindUnnamedAggColumns(columns []string) []string {
	bad := make([]string, 0)
	for _, col := range columns {
		if col == "" || aggFuncRe.MatchString(strings.TrimSpace(col)) {
			bad = append(bad, col)
		}
	}
	return bad
}
