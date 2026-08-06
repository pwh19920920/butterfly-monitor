package job

import (
	"fmt"
	"math"
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

	cur := now.Format("15:04:05")
	var msgParts []string
	level := int32(0)

	for _, group := range params {
		if len(group.EffectTimes) == 2 && (cur < group.EffectTimes[0] || cur > group.EffectTimes[1]) {
			continue
		}
		if len(group.Rules) == 0 {
			continue
		}

		// 收集每条规则的判定结果，再按 And/Or 聚合
		groupAnd := group.Relation == entity.MonitorAlertCheckParamsRelationAnd
		var hasSkip, hasFail, hasPass bool
		groupMses := make([]string, 0, len(group.Rules))

		for _, rule := range group.Rules {
			verdict, diff := evalRuleVerdict(rule, sample, real)
			switch verdict {
			case verdictSkip:
				hasSkip = true
			case verdictFail:
				hasFail = true
			case verdictPass:
				hasPass = true
				msg := formatHitRule(rule, real, sample, diff, durationSec)
				if !containsStr(groupMses, msg) {
					groupMses = append(groupMses, msg)
				}
			}
		}

		// And：全部评估且全部命中才算命中；Or：任一命中即命中
		groupHit := false
		if groupAnd {
			groupHit = !hasSkip && !hasFail && hasPass
		} else {
			groupHit = hasPass
		}
		if !groupHit {
			continue
		}

		if group.Level != nil {
			level = int32(*group.Level)
		}
		for _, part := range groupMses {
			if !containsStr(msgParts, part) {
				msgParts = append(msgParts, part)
			}
		}
		break // 组间 OR：首个命中组即打断
	}

	return len(msgParts) > 0, level, strings.Join(msgParts, "；")
}

// ruleVerdict 单条规则的评估结果
type ruleVerdict int

const (
	verdictSkip ruleVerdict = iota // 无法评估（缺样本 / 样本为 0）
	verdictFail                    // 评估完成但未命中阈值
	verdictPass                    // 命中阈值
)

// evalRuleVerdict 评估单条规则，返回三态判定。
// 依赖样本的类型在内部已做 nil / 零值守卫；实时数值类型不取 sample。
func evalRuleVerdict(rule entity.MonitorAlertCheckParamsItem, sample *float64, real float64) (ruleVerdict, float64) {
	needSample := rule.ValueType == entity.MonitorAlertCheckParamsValueTypePercent ||
		rule.ValueType == entity.MonitorAlertCheckParamsValueTypeValue
	if needSample && sample == nil {
		return verdictSkip, 0
	}

	var diff float64
	switch rule.ValueType {
	case entity.MonitorAlertCheckParamsValueTypePercent:
		if sample == nil {
			return verdictSkip, 0
		}
		s := *sample
		if s == 0 {
			return verdictSkip, 0
		}
		diff = (real - s) * 100 / s
	case entity.MonitorAlertCheckParamsValueTypeValue:
		if sample == nil {
			return verdictSkip, 0
		}
		diff = real - *sample
	default:
		diff = real
	}

	absoluteMode := rule.ValueType == entity.MonitorAlertCheckParamsValueTypeAbsoluteValue
	if compare(rule.CompareType, diff, rule.Value, absoluteMode) {
		return verdictPass, diff
	}
	return verdictFail, diff
}

// rulesNeedSample 判断规则中是否存在依赖样本基线的比较类型（样本差百分比/样本差阈值比较）。
// 用于样本缺失时决定是否保守拦截：实时数值比较规则不需要样本，不受样本缺失影响。
func rulesNeedSample(params []entity.MonitorAlertCheckParams) bool {
	for _, g := range params {
		for _, r := range g.Rules {
			if r.ValueType == entity.MonitorAlertCheckParamsValueTypePercent ||
				r.ValueType == entity.MonitorAlertCheckParamsValueTypeValue {
				return true
			}
		}
	}
	return false
}

// formatHitRule 生成命中规则文案（hitRule）
// 1 样本差阈值百分比：实时数值x1，样本数值y1，样本差百分比xy%，高于/低于样本阈值xx%，已持续发生N秒
// 2 实时数值比较：实时数值xx，高于/低于阈值yy，已持续发生N秒
// 3 样本差阈值比较：实时数值x1，样本数值y1，样本差xy，高于/低于样本阈值差xx，已持续发生N秒
func formatHitRule(rule entity.MonitorAlertCheckParamsItem, real float64, sample *float64, diff float64, durationSec int64) string {
	cmp := rule.CompareType.GetTransferMsg()
	switch rule.ValueType {
	case entity.MonitorAlertCheckParamsValueTypePercent:
		if sample == nil {
			// 防御：样本缺失时回退为实时数值比较格式
			return fmt.Sprintf(
				"实时数值%.4g，%s阈值%.4g%%，已持续发生%d秒",
				real, cmp, rule.Value, durationSec,
			)
		}
		return fmt.Sprintf(
			"实时数值%.4g，样本数值%.4g，样本差百分比%.2f%%，%s样本阈值%.4g%%，已持续发生%d秒",
			real, *sample, diff, cmp, rule.Value, durationSec,
		)
	case entity.MonitorAlertCheckParamsValueTypeValue:
		if sample == nil {
			// 防御：样本缺失时回退为实时数值比较格式
			return fmt.Sprintf(
				"实时数值%.4g，%s阈值%.4g，已持续发生%d秒",
				real, cmp, rule.Value, durationSec,
			)
		}
		return fmt.Sprintf(
			"实时数值%.4g，样本数值%.4g，样本差%.4g，%s样本阈值差%.4g，已持续发生%d秒",
			real, *sample, diff, cmp, rule.Value, durationSec,
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

// averageSampleValues 样本点聚合（抗系统抖动：天级同时段池中 1~2 天离群）。
// 调用方已保证 len(vals)>0 才调用；本函数始终返回可写值，保证界面样本线连续，
// 不会因「少样本」拒绝出数。
//
// 规则：
//   - 先丢掉 NaN/Inf（不可用点）；保留 0/负值，交由中位数与 MAD 处理（0 常为抖动日）
//   - 有效点 1 个：原样返回（单点也写，保证曲线不断）
//   - 有效点 2~4：中位数（替代原裸均值，单点 0/尖刺不再直接拉偏）
//   - 有效点 >=5：MAD 剔除离群后再中位数；剔除后过少则回退全量中位数（仍写出）
//   - 若全部为 NaN/Inf：回退 0（极端脏数据）
func averageSampleValues(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}

	clean := filterFinite(vals)
	if len(clean) == 0 {
		return 0
	}
	if len(clean) == 1 {
		return clean[0]
	}
	// 少样本也写：中位数替代算术平均
	if len(clean) < 5 {
		return medianFloat64(clean)
	}

	inliers := madInliers(clean, 3.0)
	// 剔除后过少：不放弃写点，回退全量中位数，界面仍有线
	if len(inliers) < 3 {
		return medianFloat64(clean)
	}
	return medianFloat64(inliers)
}

// filterFinite 仅去掉 NaN/Inf；保留 0 与负值，由 MAD/中位数处理离群。
func filterFinite(vals []float64) []float64 {
	out := make([]float64, 0, len(vals))
	for _, v := range vals {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			continue
		}
		out = append(out, v)
	}
	return out
}

// medianFloat64 中位数；不修改入参。
func medianFloat64(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sorted := append([]float64(nil), vals...)
	sort.Float64s(sorted)
	n := len(sorted)
	if n%2 == 1 {
		return sorted[n/2]
	}
	return (sorted[n/2-1] + sorted[n/2]) / 2
}

// madInliers 用 MAD 保留非离群点。
// 阈值 = k * 1.4826 * MAD（k=3 约等价 3σ）；MAD=0 表示几乎无离散，原样返回。
func madInliers(vals []float64, k float64) []float64 {
	if len(vals) < 3 {
		return append([]float64(nil), vals...)
	}
	med := medianFloat64(vals)
	devs := make([]float64, len(vals))
	for i, v := range vals {
		d := v - med
		if d < 0 {
			d = -d
		}
		devs[i] = d
	}
	mad := medianFloat64(devs)
	if mad == 0 {
		return append([]float64(nil), vals...)
	}
	threshold := k * 1.4826 * mad
	inliers := make([]float64, 0, len(vals))
	for _, v := range vals {
		d := v - med
		if d < 0 {
			d = -d
		}
		if d <= threshold {
			inliers = append(inliers, v)
		}
	}
	return inliers
}

// samplePoint 带来源日偏移的样本原料点（day tag = 采集时投射的未来天数，默认 1~8）
type samplePoint struct {
	Value     float64
	DayOffset int // 采集时投射的 day tag 偏移；0 表示未知
}

// filterSpecialSourceDays 丢掉「来源自然日」落在特殊日的原料点。
// 来源日 = cellEnd 往前 DayOffset 天（与 buildCollectPoints 的 day tag 语义一致）。
// isSpecial 为 nil 时原样返回。
func filterSpecialSourceDays(points []samplePoint, cellEnd time.Time, isSpecial func(time.Time) bool) []float64 {
	out := make([]float64, 0, len(points))
	for _, p := range points {
		if isSpecial != nil && p.DayOffset > 0 {
			src := cellEnd.AddDate(0, 0, -p.DayOffset)
			if isSpecial(src) {
				continue
			}
		}
		out = append(out, p.Value)
	}
	return out
}

// promoNeedSample 规则是否依赖样本基线（样本差百分比 / 样本差阈值比较）。
// 实时数值比较（AbsoluteValue）不依赖样本。
func promoNeedSample(rule entity.MonitorAlertCheckParamsItem) bool {
	return rule.ValueType == entity.MonitorAlertCheckParamsValueTypePercent ||
		rule.ValueType == entity.MonitorAlertCheckParamsValueTypeValue
}

// promoScaleFor 波动日下该条规则是否放大阈值，返回倍数（ok=false 表示不放大）。
// 放大方向 = 波动日偏移方向：
//   - peak（高峰）：上偏（Gt/Egt）放大，样本差/实时数值均可（量级整体抬高）
//   - trough（低谷）：样本差类下偏（Lt/Elt）放大（幅度 ×troughRatio）；
//     实时数值下偏不动——绝对值门槛表达不了「相对低谷的跌幅」，放大反而掩盖真故障
func promoScaleFor(dayType entity.VolatilityDayType, rule entity.MonitorAlertCheckParamsItem, peakRatio, troughRatio float64) (float64, bool) {
	switch dayType {
	case entity.VolatilityDayTypePeak:
		if rule.CompareType == entity.MonitorAlertCheckParamsCompareTypeGt ||
			rule.CompareType == entity.MonitorAlertCheckParamsCompareTypeEgt {
			return peakRatio, peakRatio > 1
		}
	case entity.VolatilityDayTypeTrough:
		if promoNeedSample(rule) &&
			(rule.CompareType == entity.MonitorAlertCheckParamsCompareTypeLt ||
				rule.CompareType == entity.MonitorAlertCheckParamsCompareTypeElt) {
			return troughRatio, troughRatio > 1
		}
	}
	return 0, false
}

// applyPromoAlertRatio 对本轮判定用的规则副本按波动日类型放大阈值（不写回 DB）。
// 触发前提：调用方已确认是敏感任务命中波动日。
// peak 放大上偏（Gt/Egt）；trough 放大样本差类下偏（Lt/Elt）。倍数<=1 时不放大。
func applyPromoAlertRatio(params []entity.MonitorAlertCheckParams, dayType entity.VolatilityDayType, peakRatio, troughRatio float64) []entity.MonitorAlertCheckParams {
	if len(params) == 0 || (peakRatio <= 1 && troughRatio <= 1) {
		return params
	}
	out := make([]entity.MonitorAlertCheckParams, len(params))
	for i, g := range params {
		out[i] = g
		if len(g.Rules) == 0 {
			continue
		}
		rules := make([]entity.MonitorAlertCheckParamsItem, len(g.Rules))
		copy(rules, g.Rules)
		for j := range rules {
			if ratio, ok := promoScaleFor(dayType, rules[j], peakRatio, troughRatio); ok {
				rules[j].Value = rules[j].Value * ratio
			}
		}
		out[i].Rules = rules
	}
	return out
}
