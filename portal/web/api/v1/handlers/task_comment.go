// Copyright 2021 CloudJ Company Limited. All rights reserved.

package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type TaskComment struct {
	ctrl.GinController
}

// Create 创建作业评论
// @Summary 创建作业评论
// @Description 创建作业评论
// @Tags 环境
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param taskId path string true "作业ID"
// @Param data body forms.CreateTaskCommentForm true "作业评论信息"
// @Success 200 {object} models.TaskComment
// @Router /tasks/{taskId}/comments [post]
func (TaskComment) Create(c *ctx.GinRequest) {
	form := &forms.CreateTaskCommentForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateTaskComment(c.Service(), form))
}

// Search 查询作业评论列表
// @Summary 查询作业评论列表
// @Description 查询作业评论列表
// @Tags 环境
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param taskId path string true "作业ID"
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]models.TaskComment}}
// @Router /tasks/{taskId}/comments [get]
func (TaskComment) Search(c *ctx.GinRequest) {
	form := &forms.SearchTaskCommentForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchTaskComment(c.Service(), form))
}
