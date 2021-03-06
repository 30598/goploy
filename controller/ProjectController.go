package controller

import (
	"bytes"
	"database/sql"
	"github.com/zhenorzz/goploy/core"
	"github.com/zhenorzz/goploy/model"
	"github.com/zhenorzz/goploy/utils"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

// Project struct
type Project Controller

// GetList -
func (project Project) GetList(gp *core.Goploy) *core.Response {
	type RespData struct {
		Projects model.Projects `json:"list"`
	}
	pagination, err := model.PaginationFrom(gp.URLQuery)
	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	projectName := gp.URLQuery.Get("projectName")
	projectList, err := model.Project{NamespaceID: gp.Namespace.ID, UserID: gp.UserInfo.ID, Name: projectName}.GetList(pagination)
	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	return &core.Response{Data: RespData{Projects: projectList}}
}

// GetTotal -
func (project Project) GetTotal(gp *core.Goploy) *core.Response {
	type RespData struct {
		Total int64 `json:"total"`
	}
	var total int64
	var err error
	projectName := gp.URLQuery.Get("projectName")
	total, err = model.Project{NamespaceID: gp.Namespace.ID, UserID: gp.UserInfo.ID, Name: projectName}.GetTotal()
	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	return &core.Response{Data: RespData{Total: total}}
}

// GetRemoteBranchList -
func (project Project) GetRemoteBranchList(gp *core.Goploy) *core.Response {
	type RespData struct {
		Branch []string `json:"branch"`
	}

	url := gp.URLQuery.Get("url")
	cmd := exec.Command("git", "ls-remote", "-h", url)
	var cmdOutbuf, cmdErrbuf bytes.Buffer
	cmd.Stdout = &cmdOutbuf
	cmd.Stderr = &cmdErrbuf
	if err := cmd.Run(); err != nil {
		return &core.Response{Code: core.Error, Message: cmdErrbuf.String()}
	}
	var branch []string
	for _, branchWithSha := range strings.Split(cmdOutbuf.String(), "\n") {
		if len(branchWithSha) != 0 {
			branchWithShaSlice := strings.Fields(branchWithSha)
			branchWithHead := branchWithShaSlice[len(branchWithShaSlice)-1]
			branchWithHeadSlice := strings.Split(branchWithHead, "/")
			branch = append(branch, branchWithHeadSlice[len(branchWithHeadSlice)-1])
		}
	}

	return &core.Response{Data: RespData{Branch: branch}}
}

// GetBindServerList -
func (project Project) GetBindServerList(gp *core.Goploy) *core.Response {
	type RespData struct {
		ProjectServers model.ProjectServers `json:"list"`
	}
	id, err := strconv.ParseInt(gp.URLQuery.Get("id"), 10, 64)
	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	projectServers, err := model.ProjectServer{ProjectID: id}.GetBindServerListByProjectID()
	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	return &core.Response{Data: RespData{ProjectServers: projectServers}}
}

// GetBindUserList -
func (project Project) GetBindUserList(gp *core.Goploy) *core.Response {
	type RespData struct {
		ProjectUsers model.ProjectUsers `json:"list"`
	}
	id, err := strconv.ParseInt(gp.URLQuery.Get("id"), 10, 64)
	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	projectUsers, err := model.ProjectUser{ProjectID: id, NamespaceID: gp.Namespace.ID}.GetBindUserListByProjectID()
	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	return &core.Response{Data: RespData{ProjectUsers: projectUsers}}
}

// Add project
func (project Project) Add(gp *core.Goploy) *core.Response {
	type ReqData struct {
		Name                  string  `json:"name" validate:"required"`
		URL                   string  `json:"url" validate:"required"`
		Path                  string  `json:"path" validate:"required"`
		Environment           string  `json:"Environment" validate:"required"`
		Branch                string  `json:"branch" validate:"required"`
		SymlinkPath           string  `json:"symlinkPath"`
		AfterPullScriptMode   string  `json:"afterPullScriptMode"`
		AfterPullScript       string  `json:"afterPullScript"`
		AfterDeployScriptMode string  `json:"afterDeployScriptMode"`
		AfterDeployScript     string  `json:"afterDeployScript"`
		RsyncOption           string  `json:"rsyncOption"`
		ServerIDs             []int64 `json:"serverIds"`
		UserIDs               []int64 `json:"userIds"`
		NotifyType            uint8   `json:"notifyType"`
		NotifyTarget          string  `json:"notifyTarget"`
	}
	var reqData ReqData
	if err := verify(gp.Body, &reqData); err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}

	if _, err := utils.ParseCommandLine(reqData.RsyncOption); err != nil {
		return &core.Response{Code: core.Error, Message: "Invalid rsync option format"}
	}

	_, err := model.Project{Name: reqData.Name}.GetDataByName()
	if err != sql.ErrNoRows {
		return &core.Response{Code: core.Error, Message: "The project name is already exist"}
	}

	projectID, err := model.Project{
		NamespaceID:           gp.Namespace.ID,
		Name:                  reqData.Name,
		URL:                   reqData.URL,
		Path:                  reqData.Path,
		SymlinkPath:           reqData.SymlinkPath,
		Environment:           reqData.Environment,
		Branch:                reqData.Branch,
		AfterPullScriptMode:   reqData.AfterPullScriptMode,
		AfterPullScript:       reqData.AfterPullScript,
		AfterDeployScriptMode: reqData.AfterDeployScriptMode,
		AfterDeployScript:     reqData.AfterDeployScript,
		RsyncOption:           reqData.RsyncOption,
		NotifyType:            reqData.NotifyType,
		NotifyTarget:          reqData.NotifyTarget,
	}.AddRow()

	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	projectServersModel := model.ProjectServers{}
	for _, serverID := range reqData.ServerIDs {
		projectServerModel := model.ProjectServer{
			ProjectID: projectID,
			ServerID:  serverID,
		}
		projectServersModel = append(projectServersModel, projectServerModel)
	}

	if err := projectServersModel.AddMany(); err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	projectUsersModel := model.ProjectUsers{}
	for _, userID := range reqData.UserIDs {
		projectUserModel := model.ProjectUser{
			ProjectID: projectID,
			UserID:    userID,
		}
		projectUsersModel = append(projectUsersModel, projectUserModel)
	}

	namespaceUsers, err := model.NamespaceUser{NamespaceID: gp.Namespace.ID}.GetAllGteManagerByNamespaceID()
	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}

	for _, namespaceUser := range namespaceUsers {
		projectUserModel := model.ProjectUser{
			ProjectID: projectID,
			UserID:    namespaceUser.UserID,
		}
		projectUsersModel = append(projectUsersModel, projectUserModel)
	}

	if err := projectUsersModel.AddMany(); err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	go repoCreate(projectID)
	return &core.Response{}
}

// Edit project
func (project Project) Edit(gp *core.Goploy) *core.Response {
	type ReqData struct {
		ID                    int64  `json:"id" validate:"gt=0"`
		Name                  string `json:"name"`
		URL                   string `json:"url"`
		Path                  string `json:"path"`
		SymlinkPath           string `json:"symlinkPath"`
		Environment           string `json:"Environment"`
		Branch                string `json:"branch"`
		AfterPullScriptMode   string `json:"afterPullScriptMode"`
		AfterPullScript       string `json:"afterPullScript"`
		AfterDeployScriptMode string `json:"afterDeployScriptMode"`
		AfterDeployScript     string `json:"afterDeployScript"`
		RsyncOption           string `json:"rsyncOption"`
		NotifyType            uint8  `json:"notifyType"`
		NotifyTarget          string `json:"notifyTarget"`
	}
	var reqData ReqData
	if err := verify(gp.Body, &reqData); err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}

	if _, err := utils.ParseCommandLine(reqData.RsyncOption); err != nil {
		return &core.Response{Code: core.Error, Message: "Invalid rsync option format"}
	}

	projectList, err := model.Project{NamespaceID: gp.Namespace.ID, Name: reqData.Name}.GetAllByName()
	if err != nil {
		if err != sql.ErrNoRows {
			return &core.Response{Code: core.Error, Message: err.Error()}
		}
	} else {
		for _, projectData := range projectList {
			if projectData.ID != reqData.ID {
				return &core.Response{Code: core.Error, Message: "The project name is already exist"}
			}
		}
	}

	projectData, err := model.Project{ID: reqData.ID}.GetData()
	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}

	err = model.Project{
		ID:                    reqData.ID,
		Name:                  reqData.Name,
		URL:                   reqData.URL,
		Path:                  reqData.Path,
		SymlinkPath:           reqData.SymlinkPath,
		Environment:           reqData.Environment,
		Branch:                reqData.Branch,
		AfterPullScriptMode:   reqData.AfterPullScriptMode,
		AfterPullScript:       reqData.AfterPullScript,
		AfterDeployScriptMode: reqData.AfterDeployScriptMode,
		AfterDeployScript:     reqData.AfterDeployScript,
		RsyncOption:           reqData.RsyncOption,
		NotifyType:            reqData.NotifyType,
		NotifyTarget:          reqData.NotifyTarget,
	}.EditRow()

	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}

	if reqData.URL != projectData.URL {
		srcPath := filepath.Join(core.RepositoryPath, projectData.Name)
		_, err := os.Stat(srcPath)
		if err == nil || os.IsNotExist(err) == false {
			repo := reqData.URL
			cmd := exec.Command("git", "remote", "set-url", "origin", repo)
			cmd.Dir = srcPath
			if err := cmd.Run(); err != nil {
				return &core.Response{Code: core.Error, Message: "Project change url fail, you can do it manually, reason: " + err.Error()}
			}
		}
	}

	if reqData.Branch != projectData.Branch {
		srcPath := filepath.Join(core.RepositoryPath, projectData.Name)
		_, err := os.Stat(srcPath)
		if err == nil || os.IsNotExist(err) == false {
			cmd := exec.Command("git", "checkout", "-f", "-B", reqData.Branch, "origin/"+reqData.Branch)
			cmd.Dir = srcPath
			if err := cmd.Run(); err != nil {
				return &core.Response{Code: core.Error, Message: "Project checkout branch fail, you can do it manually, reason: " + err.Error()}
			}
		}
	}

	// edit folder when name was change
	if reqData.Name != projectData.Name {
		srcPath := filepath.Join(core.RepositoryPath, projectData.Name)
		_, err := os.Stat(srcPath)
		if err == nil || os.IsNotExist(err) == false {
			if err := os.Rename(srcPath, filepath.Join(".", core.RepositoryPath, reqData.Name)); err != nil {
				return &core.Response{Code: core.Error, Message: "Folder rename fail, you can do it manually, reason: " + err.Error()}
			}
		}
	}

	//go repoCreate(reqData.ID)
	return &core.Response{}
}

// SetAutoDeploy -
func (project Project) SetAutoDeploy(gp *core.Goploy) *core.Response {
	type ReqData struct {
		ID         int64 `json:"id" validate:"gt=0"`
		AutoDeploy uint8 `json:"autoDeploy" validate:"gte=0"`
	}
	var reqData ReqData
	if err := verify(gp.Body, &reqData); err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	err := model.Project{
		ID:         reqData.ID,
		AutoDeploy: reqData.AutoDeploy,
	}.SetAutoDeploy()

	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	return &core.Response{}
}

// RemoveRow Project
func (project Project) Remove(gp *core.Goploy) *core.Response {
	type ReqData struct {
		ID int64 `json:"id" validate:"gt=0"`
	}
	var reqData ReqData
	if err := verify(gp.Body, &reqData); err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}

	projectData, err := model.Project{ID: reqData.ID}.GetData()
	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}

	if err := (model.Project{ID: reqData.ID}).RemoveRow(); err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}

	srcPath := path.Join(core.RepositoryPath, projectData.Name)
	if err := os.Remove(srcPath); err != nil {
		return &core.Response{Code: core.Error, Message: "Delete folder fail"}
	}

	return &core.Response{}
}

// AddServer to project
func (project Project) AddServer(gp *core.Goploy) *core.Response {
	type ReqData struct {
		ProjectID int64   `json:"projectId" validate:"gt=0"`
		ServerIDs []int64 `json:"serverIds" validate:"required"`
	}
	var reqData ReqData
	if err := verify(gp.Body, &reqData); err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	projectID := reqData.ProjectID

	projectServersModel := model.ProjectServers{}
	for _, serverID := range reqData.ServerIDs {
		projectServerModel := model.ProjectServer{
			ProjectID: projectID,
			ServerID:  serverID,
		}
		projectServersModel = append(projectServersModel, projectServerModel)
	}

	if err := projectServersModel.AddMany(); err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}

	}
	return &core.Response{}
}

// AddUser to project
func (project Project) AddUser(gp *core.Goploy) *core.Response {
	type ReqData struct {
		ProjectID int64   `json:"projectId" validate:"gt=0"`
		UserIDs   []int64 `json:"userIds" validate:"required"`
	}
	var reqData ReqData
	if err := verify(gp.Body, &reqData); err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	projectID := reqData.ProjectID

	projectUsersModel := model.ProjectUsers{}
	for _, userID := range reqData.UserIDs {
		projectUserModel := model.ProjectUser{
			ProjectID: projectID,
			UserID:    userID,
		}
		projectUsersModel = append(projectUsersModel, projectUserModel)
	}

	if err := projectUsersModel.AddMany(); err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	return &core.Response{}
}

// RemoveServer from Project
func (project Project) RemoveServer(gp *core.Goploy) *core.Response {
	type ReqData struct {
		ProjectServerID int64 `json:"projectServerId" validate:"gt=0"`
	}
	var reqData ReqData
	if err := verify(gp.Body, &reqData); err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}

	if err := (model.ProjectServer{ID: reqData.ProjectServerID}).DeleteRow(); err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	return &core.Response{}
}

// RemoveUser from Project
func (project Project) RemoveUser(gp *core.Goploy) *core.Response {
	type ReqData struct {
		ProjectUserID int64 `json:"projectUserId" validate:"gt=0"`
	}
	var reqData ReqData
	if err := verify(gp.Body, &reqData); err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}

	}

	if err := (model.ProjectUser{ID: reqData.ProjectUserID}).DeleteRow(); err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	return &core.Response{}
}

// GetTaskList -
func (project Project) GetTaskList(gp *core.Goploy) *core.Response {
	type RespData struct {
		ProjectTask model.ProjectTasks `json:"projectTaskList"`
		Pagination  model.Pagination   `json:"pagination"`
	}
	pagination, err := model.PaginationFrom(gp.URLQuery)
	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	id, err := strconv.ParseInt(gp.URLQuery.Get("id"), 10, 64)
	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	projectTaskList, pagination, err := model.ProjectTask{ProjectID: id}.GetListByProjectID(pagination)

	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}
	return &core.Response{Data: RespData{ProjectTask: projectTaskList, Pagination: pagination}}
}

// AddTask to project
func (project Project) AddTask(gp *core.Goploy) *core.Response {
	type ReqData struct {
		ProjectID int64  `json:"projectId" validate:"gt=0"`
		CommitID  string `json:"commitId" validate:"len=40"`
		Date      string `json:"date" validate:"required"`
	}
	var reqData ReqData
	if err := verify(gp.Body, &reqData); err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}

	id, err := model.ProjectTask{
		ProjectID: reqData.ProjectID,
		CommitID:  reqData.CommitID,
		Date:      reqData.Date,
		Creator:   gp.UserInfo.Name,
		CreatorID: gp.UserInfo.ID,
	}.AddRow()

	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}

	}
	type RespData struct {
		ID int64 `json:"id"`
	}
	return &core.Response{Data: RespData{ID: id}}
}

// EditTask from project
func (project Project) EditTask(gp *core.Goploy) *core.Response {
	type ReqData struct {
		ID       int64  `json:"id" validate:"gt=0"`
		CommitID string `json:"commitId" validate:"len=40"`
		Date     string `json:"date" validate:"required"`
	}
	var reqData ReqData
	if err := verify(gp.Body, &reqData); err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}

	err := model.ProjectTask{
		ID:       reqData.ID,
		CommitID: reqData.CommitID,
		Date:     reqData.Date,
		Editor:   gp.UserInfo.Name,
		EditorID: gp.UserInfo.ID,
	}.EditRow()

	if err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}

	}

	return &core.Response{}
}

// RemoveTask from project
func (project Project) RemoveTask(gp *core.Goploy) *core.Response {
	type ReqData struct {
		ID int64 `json:"id" validate:"gt=0"`
	}
	var reqData ReqData
	if err := verify(gp.Body, &reqData); err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}

	if err := (model.ProjectTask{ID: reqData.ID}).RemoveRow(); err != nil {
		return &core.Response{Code: core.Error, Message: err.Error()}
	}

	return &core.Response{}
}

// repoCreate -
func repoCreate(projectID int64) {
	project, err := model.Project{ID: projectID}.GetData()
	if err != nil {
		core.Log(core.TRACE, "The project does not exist, projectID:"+strconv.FormatInt(projectID, 10))
		return
	}
	srcPath := core.RepositoryPath + project.Name
	if _, err := os.Stat(srcPath); err != nil {
		if err := os.RemoveAll(srcPath); err != nil {
			core.Log(core.TRACE, "The project fail to remove, projectID:"+strconv.FormatInt(project.ID, 10))
			return
		}
		repo := project.URL
		cmd := exec.Command("git", "clone", repo, srcPath)
		var out bytes.Buffer
		cmd.Stdout = &out

		if err := cmd.Run(); err != nil {
			core.Log(core.ERROR, "The project fail to initialize, projectID:"+strconv.FormatInt(project.ID, 10)+err.Error())
			return
		}

		if project.Branch != "master" {
			checkout := exec.Command("git", "checkout", "-b", project.Branch, "origin/"+project.Branch)
			checkout.Dir = srcPath
			var checkoutOutbuf, checkoutErrbuf bytes.Buffer
			checkout.Stdout = &checkoutOutbuf
			checkout.Stderr = &checkoutErrbuf
			if err := checkout.Run(); err != nil {
				core.Log(core.ERROR, "projectID:"+strconv.FormatInt(project.ID, 10)+checkoutErrbuf.String())
				os.RemoveAll(srcPath)
				return
			}

		}
		core.Log(core.TRACE, "The project success to initialize, projectID:"+strconv.FormatInt(project.ID, 10))

	}
	return
}
