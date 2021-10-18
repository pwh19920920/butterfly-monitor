package handler

import (
	"butterfly-monitor/src/app/domain/entity"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kirinlabs/HttpRequest"
	"github.com/thedevsaddam/gojsonq/v2"
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

	req := HttpRequest.NewRequest()
	resp, err := req.Get(task.Command)
	if err != nil || resp.StatusCode() != 200 {
		return 0, errors.New(fmt.Sprintf("请求url: %v 发生错误", task.Command))
	}

	body, err := resp.Body()
	if err != nil {
		return 0, err
	}

	result, err := gojsonq.New().FromString(string(body)).FindR(params.ResultFieldPath)
	if err != nil || result == nil {
		return nil, errors.New("请求错误, 或者取不到结果")
	}
	return result.Float64()
}
