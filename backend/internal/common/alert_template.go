package common

import (
	"bytes"
	"text/template"
)

// AlertTemplateItem 单条告警渲染项
type AlertTemplateItem struct {
	TaskName string
	HitRule  string
}

// RenderAlertTemplateMulti 渲染多条聚合告警
// 模板可用字段：{{range .items}}{{.TaskName}}{{.HitRule}}{{end}} 遍历各条告警；
// {{range .relationTaskNames}}{{.}}{{end}} 遍历关联告警文案（每条一个字符串）
func RenderAlertTemplateMulti(tplStr string, items []AlertTemplateItem, relationTaskNames []string) (string, error) {
	tpl, err := template.New("alert").Parse(tplStr)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tpl.Execute(&buf, map[string]interface{}{
		"items":             items,
		"relationTaskNames": relationTaskNames,
	})
	return buf.String(), err
}

// RenderAlertTemplate 渲染报警模板（单条，兼容旧调用方）
// relationTaskNames 为关联告警文案列表（可为空）；模板字段 RelationTaskNames，不应包含当前任务自身
func RenderAlertTemplate(tplStr, taskName, hitRule, relationTaskNames string) (string, error) {
	var rel []string
	if relationTaskNames != "" {
		rel = []string{relationTaskNames}
	}
	return RenderAlertTemplateMulti(tplStr, []AlertTemplateItem{{
		TaskName: taskName,
		HitRule:  hitRule,
	}}, rel)
}
