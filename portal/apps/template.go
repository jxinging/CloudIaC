// Copyright 2021 CloudJ Company Limited. All rights reserved.

package apps

import (
	"cloudiac/configs"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/libs/page"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"cloudiac/portal/services/vcsrv"
	"cloudiac/utils"
	"fmt"
	"net/http"
)

type SearchTemplateResp struct {
	CreatedAt         models.Time `json:"createdAt"` // 创建时间
	UpdatedAt         models.Time `json:"updatedAt"` // 更新时间
	Id                models.Id   `json:"id"`
	Name              string      `json:"name"`
	Description       string      `json:"description"`
	ActiveEnvironment int         `json:"activeEnvironment"`
	RepoRevision      string      `json:"repoRevision"`
	Creator           string      `json:"creator"`
	RepoId            string      `json:"repoId"`
	VcsId             string      `json:"vcsId"`
	RepoAddr          string      `json:"repoAddr"`
	TplType           string      `json:"tplType" `
	RepoFullName      string      `json:"repoFullName"`
	NewRepoAddr       string      `json:"newRepoAddr"`
	VcsAddr           string      `json:"vcsAddr"`
}

func getRepoAddr(vcsId models.Id, query *db.Session, repoId string) (string, error) {
	vcs, err := services.QueryVcsByVcsId(vcsId, query)
	if err != nil {
		return "", err
	}
	vcsIface, er := vcsrv.GetVcsInstance(vcs)
	if er != nil {
		return "", er
	}
	repo, er := vcsIface.GetRepo(repoId)
	if er != nil {
		return "", er
	}
	repoAddr, er := vcsrv.GetRepoAddress(repo)
	if er != nil {
		return "", er
	}
	return repoAddr, nil
}

func getRepo(vcsId models.Id, query *db.Session, repoId string) (*vcsrv.Projects, error) {
	vcs, err := services.QueryVcsByVcsId(vcsId, query)
	if err != nil {
		return nil, err
	}
	vcsIface, er := vcsrv.GetVcsInstance(vcs)
	if er != nil {
		return nil, er
	}
	repo, er := vcsIface.GetRepo(repoId)
	if er != nil {
		return nil, er
	}

	p, err := repo.FormatRepoSearch()
	if err != nil {
		return nil, err
	}
	return p, nil
}

func CreateTemplate(c *ctx.ServiceContext, form *forms.CreateTemplateForm) (*models.Template, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create template %s", form.Name))

	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()
	template, err := services.CreateTemplate(tx, models.Template{
		Name:         form.Name,
		OrgId:        c.OrgId,
		Description:  form.Description,
		VcsId:        form.VcsId,
		RepoId:       form.RepoId,
		RepoFullName: form.RepoFullName,
		// 云模板的 repoAddr 和 repoToken 可以为空，若为空则会在创建任务时自动查询 vcs 获取相应值
		RepoAddr:     "",
		RepoToken:    "",
		RepoRevision: form.RepoRevision,
		CreatorId:    c.UserId,
		Workdir:      form.Workdir,
		Playbook:     form.Playbook,
		PlayVarsFile: form.PlayVarsFile,
		TfVarsFile:   form.TfVarsFile,
		TfVersion:    form.TfVersion,
	})

	if err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error create template, err %s", err)
		if err.Code() == e.TemplateAlreadyExists {
			return nil, e.New(err.Code(), err.Err(), http.StatusBadRequest)
		}
		return nil, err
	}

	// 创建模板与项目的关系
	if err := services.CreateTemplateProject(tx, form.ProjectId, template.Id); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	{
		updateVarsForm := forms.UpdateObjectVarsForm{
			Scope:     consts.ScopeTemplate,
			ObjectId:  template.Id,
			Variables: form.Variables,
		}
		if _, er := updateObjectVars(c, tx, &updateVarsForm); er != nil {
			_ = tx.Rollback()
			return nil, er
		}
	}

	// 创建变量组与实例的关系
	if err := services.BatchUpdateRelationship(tx, form.VarGroupIds, form.DelVarGroupIds, consts.ScopeTemplate, template.Id.String()); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error commit create template, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	return template, nil
}

func UpdateTemplate(c *ctx.ServiceContext, form *forms.UpdateTemplateForm) (*models.Template, e.Error) {
	c.AddLogField("action", fmt.Sprintf("update template %s", form.Id))

	tpl, err := services.GetTemplateById(c.DB(), form.Id)
	if err != nil {
		return nil, e.New(e.TemplateNotExists, err, http.StatusBadRequest)
	}

	// 根据云模板ID, 组织ID查询该云模板是否属于该组织
	if tpl.OrgId != c.OrgId {
		return nil, e.New(e.TemplateNotExists, http.StatusForbidden, fmt.Errorf("the organization does not have permission to delete the current template"))
	}
	attrs := models.Attrs{}
	if form.HasKey("name") {
		attrs["name"] = form.Name
	}

	if form.HasKey("description") {
		attrs["description"] = form.Description
	}
	if form.HasKey("playbook") {
		attrs["playbook"] = form.Playbook
	}
	if form.HasKey("status") {
		attrs["status"] = form.Status
	}
	if form.HasKey("workdir") {
		attrs["workdir"] = form.Workdir
	}
	if form.HasKey("tfVarsFile") {
		attrs["tfVarsFile"] = form.TfVarsFile
	}
	if form.HasKey("playVarsFile") {
		attrs["playVarsFile"] = form.PlayVarsFile
	}
	if form.HasKey("tfVersion") {
		attrs["tfVersion"] = form.TfVersion
	}
	if form.HasKey("repoRevision") {
		attrs["repoRevision"] = form.RepoRevision
	}
	if form.HasKey("vcsId") && form.HasKey("repoId") && form.HasKey("repoFullName") {
		attrs["vcsId"] = form.VcsId
		attrs["repoId"] = form.RepoId
		attrs["repoFullName"] = form.RepoFullName
	}
	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()
	if tpl, err = services.UpdateTemplate(tx, form.Id, attrs); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if form.HasKey("projectId") {
		if err := services.DeleteTemplateProject(tx, form.Id); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		if err := services.CreateTemplateProject(tx, form.ProjectId, form.Id); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}
	if form.HasKey("variables") {
		updateVarsForm := forms.UpdateObjectVarsForm{
			Scope:     consts.ScopeTemplate,
			ObjectId:  form.Id,
			Variables: form.Variables,
		}
		if _, er := updateObjectVars(c, tx, &updateVarsForm); er != nil {
			_ = tx.Rollback()
			return nil, er
		}
	}

	if form.HasKey("varGroupIds") || form.HasKey("delVarGroupIds") {
		// 创建变量组与实例的关系
		if err := services.BatchUpdateRelationship(tx, form.VarGroupIds, form.DelVarGroupIds, consts.ScopeTemplate, form.Id.String()); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error commit update template, err %s", err)
		return nil, e.New(e.DBError, err)
	}
	return tpl, err
}

func DeleteTemplate(c *ctx.ServiceContext, form *forms.DeleteTemplateForm) (interface{}, e.Error) {
	c.AddLogField("action", fmt.Sprintf("delete template %s", form.Id))
	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()
	// 根据ID 查询云模板是否存在
	tpl, err := services.GetTemplateById(tx, form.Id)
	if err != nil && err.Code() == e.TemplateNotExists {
		return nil, e.New(err.Code(), err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get template by id, err %v", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}
	// 根据云模板ID, 组织ID查询该云模板是否属于该组织
	if tpl.OrgId != c.OrgId {
		return nil, e.New(e.TemplateNotExists, http.StatusForbidden, fmt.Errorf("The organization does not have permission to delete the current template"))
	}

	// 查询模板是否有活跃环境
	if ok, err := services.QueryActiveEnv(tx.Where("tpl_id = ?", form.Id)).Exists(); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	} else if ok {
		return nil, e.New(e.TemplateActiveEnvExists, http.StatusMethodNotAllowed,
			fmt.Errorf("The cloud template cannot be deleted because there is an active environment"))
	}

	// 根据ID 删除云模板
	if err := services.DeleteTemplate(tx, tpl.Id); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error commit del template, err %s", err)
		return nil, e.New(e.DBError, err)
	}
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error commit del template, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	return nil, nil

}

type TemplateDetailResp struct {
	*models.Template
	Variables   []models.Variable `json:"variables"`
	ProjectList []models.Id       `json:"projectId"`
}

func TemplateDetail(c *ctx.ServiceContext, form *forms.DetailTemplateForm) (*TemplateDetailResp, e.Error) {
	tpl, err := services.GetTemplateById(c.DB(), form.Id)
	if err != nil && err.Code() == e.TemplateNotExists {
		return nil, e.New(err.Code(), err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get template by id, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}
	project_ids, err := services.QueryProjectByTplId(c.DB(), form.Id)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}
	varialbeList, err := services.SearchVariableByTemplateId(c.DB(), form.Id)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	if tpl.RepoFullName == "" {
		repo, err := getRepo(tpl.VcsId, c.DB(), tpl.RepoId)
		if err != nil {
			return nil, e.New(e.VcsError, err)
		}
		tpl.RepoFullName = repo.FullName
	}

	tplDetail := &TemplateDetailResp{
		Template:    tpl,
		Variables:   varialbeList,
		ProjectList: project_ids,
	}
	return tplDetail, nil

}

func SearchTemplate(c *ctx.ServiceContext, form *forms.SearchTemplateForm) (tpl interface{}, err e.Error) {
	tplIdList := make([]models.Id, 0)
	if c.ProjectId != "" {
		tplIdList, err = services.QueryTplByProjectId(c.DB(), c.ProjectId)
		if err != nil {
			return nil, err
		}

		if len(tplIdList) == 0 {
			return getEmptyListResult(form)
		}
	}
	vcsIds := make([]string, 0)
	query := services.QueryTemplateByOrgId(c.DB(), form.Q, c.OrgId, tplIdList)
	p := page.New(form.CurrentPage(), form.PageSize(), query)
	templates := make([]*SearchTemplateResp, 0)
	if err := p.Scan(&templates); err != nil {
		return nil, e.New(e.DBError, err)
	}

	for _, v := range templates {
		if v.RepoAddr == "" {
			vcsIds = append(vcsIds, v.VcsId)
		}
	}

	vcsList, err := services.GetVcsListByIds(c.DB(), vcsIds)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	vcsAttr := make(map[string]models.Vcs)
	for _, v := range vcsList {
		vcsAttr[v.Id.String()] = v
	}

	for _, tpl := range templates {
		if tpl.RepoAddr == "" && tpl.RepoFullName != "" {
			if vcsAttr[tpl.VcsId].VcsType == consts.GitTypeLocal {
				tpl.RepoAddr = fmt.Sprintf("%s/%s/%s.git", utils.GetUrl(configs.Get().Portal.Address), vcsAttr[tpl.VcsId].Address, tpl.RepoFullName)
				continue
			}
			tpl.RepoAddr = fmt.Sprintf("%s/%s.git", utils.GetUrl(vcsAttr[tpl.VcsId].Address), tpl.RepoFullName)
		}
	}

	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     templates,
	}, nil
}

type TemplateChecksResp struct {
	CheckResult string `json:"CheckResult"`
	Reason      string `json:"reason"`
}

func TemplateChecks(c *ctx.ServiceContext, form *forms.TemplateChecksForm) (interface{}, e.Error) {

	// 如果云模版名称传入，校验名称是否重复.
	if form.Name != "" {
		tpl, err := services.QueryTemplateByName(c.DB().Where("id != ?", form.TemplateId), form.Name, c.OrgId)
		if tpl != nil {
			return nil, e.New(e.TemplateNameRepeat, err)
		}
		// 数据库相关错误
		if err != nil && err.Code() != e.TemplateNotExists {
			return nil, err
		}
	}
	if form.Workdir != "" {
		// 检查工作目录下.tf 文件是否存在
		searchForm := &forms.TemplateTfvarsSearchForm{
			RepoId:       form.RepoId,
			RepoRevision: form.RepoRevision,
			RepoType:     form.RepoType,
			VcsId:        form.VcsId,
			TplChecks:    true,
			Path:         form.Workdir,
		}
		results, err := VcsFileSearch(c, searchForm)
		if err != nil {
			return nil, err
		}
		if len(results.([]string)) == 0 {
			return nil, e.New(e.TemplateWorkdirError, err)
		}
	}
	return TemplateChecksResp{
		CheckResult: consts.TplTfCheckSuccess,
	}, nil
}
