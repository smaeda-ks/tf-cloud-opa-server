package lib

import "time"

type TFCloudRunTasksRequest struct {
	PayloadVersion             int       `json:"payload_version"`
	AccessToken                string    `json:"access_token"`
	Stage                      string    `json:"stage"`
	IsSpeculative              bool      `json:"is_speculative"`
	TaskResultID               string    `json:"task_result_id"`
	TaskResultEnforcementLevel string    `json:"task_result_enforcement_level"`
	TaskResultCallbackURL      string    `json:"task_result_callback_url"`
	RunID                      string    `json:"run_id"`
	RunAppURL                  string    `json:"run_app_url"`
	RunMessage                 string    `json:"run_message"`
	RunCreatedAt               time.Time `json:"run_created_at"`
	RunCreatedBy               string    `json:"run_created_by"`
	WorkspaceID                string    `json:"workspace_id"`
	WorkspaceName              string    `json:"workspace_name"`
	WorkspaceAppURL            string    `json:"workspace_app_url"`
	OrganizationName           string    `json:"organization_name"`
	PlanJSONAPIURL             string    `json:"plan_json_api_url"`
	VcsRepoURL                 string    `json:"vcs_repo_url"`
	VcsBranch                  string    `json:"vcs_branch"`
	VcsPullRequestURL          string    `json:"vcs_pull_request_url"`
	VcsCommitURL               string    `json:"vcs_commit_url"`
}

type TFCloudRunTasksResponse struct {
	Data *TFCloudRunTasksResponseData `json:"data"`
}

type TFCloudRunTasksResponseData struct {
	Type       string                             `json:"type"`
	Attributes *TFCloudRunTasksResponseAttributes `json:"attributes"`
}

type TFCloudRunTasksResponseAttributes struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Url     string `json:"-"`
}

type TFCloudComment struct {
	Data *TFCloudCommentData `json:"data"`
}

type TFCloudCommentData struct {
	Type       string                    `json:"type"`
	Attributes *TFCloudCommentAttributes `json:"attributes"`
}

type TFCloudCommentAttributes struct {
	Body string `json:"body"`
}
