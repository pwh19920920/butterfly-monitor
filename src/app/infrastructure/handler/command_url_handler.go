package handler

import (
	"butterfly-monitor/src/app/domain/entity"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/kirinlabs/HttpRequest"
)

type CommandUrlHandler struct {
}

func (urlHandler *CommandUrlHandler) ExecuteCommand(task entity.JobTask) (int64, error) {
	req := HttpRequest.NewRequest()
	resp, err := req.Get(task.Command)
	if err != nil || resp.StatusCode() != 200 {
		return 0, errors.New(fmt.Sprintf("请求url: %v 发生错误", task.Command))
	}

	body, err := resp.Body()
	if err != nil {
		return 0, err
	}

	binBuf := bytes.NewBuffer(body)
	var result int64
	err = binary.Read(binBuf, binary.BigEndian, &result)
	return result, err
}
