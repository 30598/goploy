package controller

import (
	"github.com/zhenorzz/goploy/core"
	"github.com/zhenorzz/goploy/model"
)

// Template struct
type Template Controller

// GetList -
func (Template) GetList(gp *core.Goploy) *core.Response {
	type RespData struct {
		Templates model.Templates `json:"list"`
	}
	pagination, err := model.PaginationFrom(gp.URLQuery)
	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	templateList, err := model.Template{}.GetList(pagination)
	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	return &core.Response{Data: RespData{Templates: templateList}}
}

// GetTotal -
func (Template) GetTotal(gp *core.Goploy) *core.Response {
	type RespData struct {
		Total int64 `json:"total"`
	}
	total, err := model.Template{}.GetTotal()
	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	return &core.Response{Data: RespData{Total: total}}
}

// GetOption -
func (Template) GetOption(gp *core.Goploy) *core.Response {
	type RespData struct {
		Templates model.Templates `json:"list"`
	}

	templateList, err := model.Template{}.GetAll()
	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	return &core.Response{Data: RespData{Templates: templateList}}
}

// Add template
func (Template) Add(gp *core.Goploy) *core.Response {
	type ReqData struct {
		Name         string `json:"name" validate:"required"`
		Remark       string `json:"remark"`
		PackageIDStr string `json:"packageIdStr"`
		Script       string `json:"script" validate:"required"`
	}
	type RespData struct {
		ID int64 `json:"id"`
	}

	var reqData ReqData

	if err := verify(gp.Body, &reqData); err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}

	id, err := model.Template{
		Name:         reqData.Name,
		Remark:       reqData.Remark,
		PackageIDStr: reqData.PackageIDStr,
		Script:       reqData.Script,
	}.AddRow()

	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	return &core.Response{Data: RespData{ID: id}}
}

// Edit template
func (Template) Edit(gp *core.Goploy) *core.Response {
	type ReqData struct {
		ID           int64  `json:"id" validate:"gt=0"`
		Name         string `json:"name" validate:"required"`
		Remark       string `json:"remark"`
		PackageIDStr string `json:"packageIdStr"`
		Script       string `json:"script" validate:"required"`
	}
	var reqData ReqData
	if err := verify(gp.Body, &reqData); err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	err := model.Template{
		ID:           reqData.ID,
		Name:         reqData.Name,
		Remark:       reqData.Remark,
		PackageIDStr: reqData.PackageIDStr,
		Script:       reqData.Script,
	}.EditRow()

	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	return &core.Response{}
}

// RemoveRow Template
func (Template) Remove(gp *core.Goploy) *core.Response {
	type ReqData struct {
		ID int64 `json:"id" validate:"gt=0"`
	}
	var reqData ReqData
	if err := verify(gp.Body, &reqData); err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}

	if err := (model.Template{ID: reqData.ID}).DeleteRow(); err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	return &core.Response{Message: "删除成功"}
}
