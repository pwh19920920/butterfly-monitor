package handler

import (
	"butterfly-monitor/domain/entity"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kirinlabs/HttpRequest"
	"github.com/thedevsaddam/gojsonq/v2"
	"net/url"
	"strconv"
)

type CommandUrlHandler struct {
}

type CommandUrlParams struct {
	ResultFieldPath string `json:"resultFieldPath"` // 支持对象.属性
}

func (urlHandler *CommandUrlHandler) ExecuteCommand(task entity.MonitorTask) (interface{}, error) {
	if task.ExecParams == "" {
		return 0, errors.New("执行参数有误")
	}

	var params CommandUrlParams
	err := json.Unmarshal([]byte(task.ExecParams), &params)
	if err != nil {
		return 0, err
	}

	parseResult, err := url.Parse(task.Command)
	if err != nil {
		return 0, errors.New("路径转换失败")
	}

	req := HttpRequest.NewRequest()
	resp, err := req.Get(fmt.Sprintf("%v://%v%v", parseResult.Scheme, parseResult.Host, parseResult.Path), url.PathEscape(parseResult.RawQuery))
	if err != nil || resp.StatusCode() != 200 {
		return 0, errors.New(fmt.Sprintf("请求url: %v 发生错误", task.Command))
	}

	body, err := resp.Body()
	if err != nil {
		return 0, err
	}

	result := gojsonq.New().FromString(string(body)).Find(params.ResultFieldPath)
	if nil == result {
		return nil, errors.New("请求成功, 但取不到结果")
	}

	return strconv.ParseFloat(fmt.Sprintf("%v", result), 64)
}
