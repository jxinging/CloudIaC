// Copyright 2021 CloudJ Company Limited. All rights reserved.

package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
)

type SystemStatusResp struct {
	Service  string `json:"service" form:"service" `
	Children []struct {
		ID      string
		Tags    []string `json:"tags" form:"tags" `
		Port    int      `json:"port" form:"port" `
		Address string   `json:"address" form:"address" `
		Status  string   `json:"status" form:"status" `
		Node    string   `json:"node" form:"node" `
		Notes   string   `json:"notes" form:"notes" `
		Output  string   `json:"output" form:"output" `
	}
	//Passing  uint64 `json:"passing" form:"passing" `
	//Critical uint64 `json:"critical" form:"critical" `
	//Warn     uint64 `json:"warn" form:"warn" `
}

func SystemStatusSearch() (interface{}, e.Error) {
	resp := make([]*SystemStatusResp, 0)
	serviceResp := make(map[string]*SystemStatusResp, 0)
	IdInfo, serviceStatus, serviceList, err := services.SystemStatusSearch()
	if err != nil {
		return nil, err
	}

	//构建返回值
	for _, service := range serviceList {
		serviceResp[service] = &SystemStatusResp{
			Service: service,
		}
	}

	for _, id := range IdInfo {
		serviceResp[id.Service].Children = append(serviceResp[id.Service].Children, struct {
			ID      string
			Tags    []string `json:"tags" form:"tags" `
			Port    int      `json:"port" form:"port" `
			Address string   `json:"address" form:"address" `
			Status  string   `json:"status" form:"status" `
			Node    string   `json:"node" form:"node" `
			Notes   string   `json:"notes" form:"notes" `
			Output  string   `json:"output" form:"output" `
		}{
			ID:      id.ID,
			Tags:    id.Tags,
			Port:    id.Port,
			Address: id.Address,
			Status:  serviceStatus[id.ID].Status,
			Node:    serviceStatus[id.ID].Node,
			Notes:   serviceStatus[id.ID].Notes,
			Output:  serviceStatus[id.ID].Output,
		})
	}

	for _, service := range serviceResp {
		resp = append(resp, service)
	}

	return resp, nil
}

func ConsulKVSearch(key string) (interface{}, e.Error) {
	return services.ConsulKVSearch(key)
}

func RunnerSearch() (interface{}, e.Error) {
	return services.RunnerSearch()
}

func ConsulTagUpdate(form forms.ConsulTagUpdateForm) (interface{}, e.Error) {
	//将修改后的tag存到consul中
	if err := services.ConsulKVSave(form.ServiceId, form.Tags); err != nil {
		return nil, err
	}
	//根据serviceId查询在consul中保存的数据
	agentService, err := services.ConsulServiceInfo(form.ServiceId)
	if err != nil {
		return nil, err
	}
	//重新注册
	if err := services.ConsulServiceRegistered(agentService, form.Tags); err != nil {
		return nil, err
	}
	return nil, nil
}
