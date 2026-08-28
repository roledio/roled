package models

type BuildInfo struct {
	Env            string `json:"env"`
	ProjectName    string `json:"project_name"`
	AppName        string `json:"app_name"`
	AppVersion     string `json:"app_version"`
	CommitHash     string `json:"commit_hash"`
	BuildTimestamp string `json:"build_timestamp"`
	StartTimestamp string `json:"start_timestamp"`
	Age            string `json:"age"`
}
