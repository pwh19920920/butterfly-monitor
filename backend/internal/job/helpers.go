package job

import (
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"dragonfly-monitor/internal/domain/entity"

	"github.com/sirupsen/logrus"
)

func recoverLog(name string) {
	if r := recover(); r != nil {
		buf := make([]byte, 4096)
		n := runtime.Stack(buf, false)
		logrus.Errorf("%s panic: %v\n%s", name, r, string(buf[:n]))
	}
}

// renderCommand 模板渲染：把 SQL/URL 命令中的 {{.xxx}} 占位替换为时间参数。
// 使用简单字符串替换替代 text/template 引擎，避免用户可控的 command 字段被当作模板语法解析，
// 防止未来 FuncMap 变更导致潜在的 RCE 风险。
func renderCommand(cmd string, begin, start, end time.Time) (string, error) {
	repl := map[string]string{
		"{{.beginTime}}":      begin.Format("2006-01-02 15:04:05"),
		"{{.startTime}}":      start.Format("2006-01-02 15:04:05"),
		"{{.endTime}}":        end.Format("2006-01-02 15:04:05"),
		"{{.beginTimeMilli}}": fmt.Sprintf("%d", begin.UnixMilli()),
		"{{.startTimeMilli}}": fmt.Sprintf("%d", start.UnixMilli()),
		"{{.endTimeMilli}}":   fmt.Sprintf("%d", end.UnixMilli()),
	}
	result := cmd
	for k, v := range repl {
		result = strings.ReplaceAll(result, k, v)
	}
	return result, nil
}

// evaluateRules 规则判定
// 语义：
// 1. 组与组之间固定 OR：首个规则组命中即整体命中，立即打断不再匹配后续组
// 2. 组内关系由 group.Relation 指定：Or=任一规则命中，And=全部规则命中
// 3. 不在生效时段的组跳过，不参与组间 OR
// 4. hitRule 文案只收集「首个命中组」的规则；相同文案去重
func evaluateRules(params []entity.MonitorAlertCheckParams, sample *float64, real float64, now time.Time, durationSec int64) (bool, int32, string) {
	if len(params) == 0 {
		return false, 0, ""
	}

	overallHit := false
	msgParts := make([]string, 0)
	level := int32(0)

	for _, group := range params {
		// 生效时段
		if len(group.EffectTimes) == 2 {
			cur := now.Format("15:04:05")
			if cur < group.EffectTimes[0] || cur > group.EffectTimes[1] {
				continue
			}
		}
		if len(group.Rules) == 0 {
			continue
		}

		// 组内：按 relation 聚合（默认 Or）
		groupAnd := group.Relation == entity.MonitorAlertCheckParamsRelationAnd
		groupHit := groupAnd // And 初始 true，任一失败变 false；Or 初始 false，任一成功变 true
		groupMsgs := make([]string, 0, len(group.Rules))
		anyEvaluated := false

		for _, rule := range group.Rules {
			// 三种比较：
			// 1 样本差阈值百分比：diff=(实时-样本)/样本*100
			// 2 样本差阈值比较：diff=实时-样本
			// 3 实时数值比较：diff=实时
			var diff float64
			needSample := rule.ValueType == entity.MonitorAlertCheckParamsValueTypePercent ||
				rule.ValueType == entity.MonitorAlertCheckParamsValueTypeAbsoluteValue
			if needSample && sample == nil {
				// 缺样本：该条无法判定
				// And 组：视为本条未通过；Or 组：跳过
				if groupAnd {
					groupHit = false
				}
				continue
			}
			switch rule.ValueType {
			case entity.MonitorAlertCheckParamsValueTypePercent:
				if *sample == 0 {
					if groupAnd {
						groupHit = false
					}
					continue
				}
				diff = (real - *sample) * 100 / *sample
			case entity.MonitorAlertCheckParamsValueTypeAbsoluteValue:
				diff = real - *sample
			case entity.MonitorAlertCheckParamsValueTypeValue:
				diff = real
			default:
				diff = real
			}
			// 样本差类型：低于/低于或等于 与阈值取负方向比较（波动幅度）
			// 实时数值：直接 real 与阈值比较
			absoluteMode := rule.ValueType == entity.MonitorAlertCheckParamsValueTypeValue
			ok := compare(rule.CompareType, diff, rule.Value, absoluteMode)
			anyEvaluated = true
			if ok {
				part := formatHitRule(rule, real, sample, diff, durationSec)
				if !containsStr(groupMsgs, part) {
					groupMsgs = append(groupMsgs, part)
				}
				if !groupAnd {
					groupHit = true
				}
			} else if groupAnd {
				groupHit = false
			}
		}

		// 组内没有任何可评估规则时，不视为命中
		if !anyEvaluated {
			groupHit = false
		}
		if group.Level != nil {
			// 命中组才采纳其 level；组间 OR 一旦命中即打断，故只取首个命中组的 level
			if groupHit {
				level = int32(*group.Level)
			}
		}
		if groupHit {
			overallHit = true
			// 只拼本命中组的文案，避免未命中组污染
			for _, part := range groupMsgs {
				if !containsStr(msgParts, part) {
					msgParts = append(msgParts, part)
				}
			}
			// 组间 OR：首个命中组即打断，不再匹配后续组
			break
		}
	}

	return overallHit, level, strings.Join(msgParts, "；")
}

// formatHitRule 生成命中规则文案（hitRule）
// 1 实时数值比较：实时数值xx，高于/低于阈值yy，已持续发生N秒
// 2 样本差阈值百分比：实时数值x1，样本数值y1，样本差百分比xy%，高于/低于样本阈值xx%，已持续发生N秒
// 3 样本差阈值比较：实时数值x1，样本数值y1，样本差xy，高于/低于样本阈值差xx，已持续发生N秒
func formatHitRule(rule entity.MonitorAlertCheckParamsItem, real float64, sample *float64, diff float64, durationSec int64) string {
	cmp := rule.CompareType.GetTransferMsg()
	switch rule.ValueType {
	case entity.MonitorAlertCheckParamsValueTypePercent:
		sampleVal := 0.0
		if sample != nil {
			sampleVal = *sample
		}
		return fmt.Sprintf(
			"实时数值%.4g，样本数值%.4g，样本差百分比%.2f%%，%s样本阈值%.4g%%，已持续发生%d秒",
			real, sampleVal, diff, cmp, rule.Value, durationSec,
		)
	case entity.MonitorAlertCheckParamsValueTypeAbsoluteValue:
		sampleVal := 0.0
		if sample != nil {
			sampleVal = *sample
		}
		return fmt.Sprintf(
			"实时数值%.4g，样本数值%.4g，样本差%.4g，%s样本阈值差%.4g，已持续发生%d秒",
			real, sampleVal, diff, cmp, rule.Value, durationSec,
		)
	default:
		// 实时数值比较
		return fmt.Sprintf(
			"实时数值%.4g，%s阈值%.4g，已持续发生%d秒",
			real, cmp, rule.Value, durationSec,
		)
	}
}

// compare 比较运算
// absoluteMode=true：实时数值比较，diff 与 value 直接比
// absoluteMode=false：样本差比较，"低于/低于或等于" 按 -value 判断负向波动
func compare(ct entity.MonitorAlertCheckParamsCompareType, diff, value float64, absoluteMode bool) bool {
	switch ct {
	case entity.MonitorAlertCheckParamsCompareTypeGt:
		return diff > value
	case entity.MonitorAlertCheckParamsCompareTypeLt:
		if absoluteMode {
			return diff < value
		}
		return diff < -value
	case entity.MonitorAlertCheckParamsCompareTypeEq:
		return diff == value
	case entity.MonitorAlertCheckParamsCompareTypeEgt:
		return diff >= value
	case entity.MonitorAlertCheckParamsCompareTypeElt:
		if absoluteMode {
			return diff <= value
		}
		return diff <= -value
	}
	return false
}

func containsStr(ss []string, target string) bool {
	for _, s := range ss {
		if s == target {
			return true
		}
	}
	return false
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseInt64(s string) (int64, error) {
	// 用 strconv.ParseInt 替代 fmt.Sscanf，解析更严格也更高效
	v, err := strconv.ParseInt(s, 10, 64)
	return v, err
}

// averageSampleValues 样本点聚合：
// - 点数 < 5：直接算术平均
// - 点数 >= 5：去掉最大最小后再平均
func averageSampleValues(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	if len(vals) < 5 {
		var sum float64
		for _, v := range vals {
			sum += v
		}
		return sum / float64(len(vals))
	}
	sort.Float64s(vals)
	trimmed := vals[1 : len(vals)-1]
	var sum float64
	for _, v := range trimmed {
		sum += v
	}
	return sum / float64(len(trimmed))
}
