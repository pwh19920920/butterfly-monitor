package types

import "github.com/pwh19920920/butterfly/pkg/response"

type AlertGroupQueryRequest struct {
	response.RequestPaging
	Name string `form:"name"`
}

type AlertGroupSaveRequest struct {
	Id      int64    `json:"id,string"`
	Name    string   `json:"name"`
	UserIds []string `json:"userIds"`
}
