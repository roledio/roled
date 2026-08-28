package models

type RequestLog struct {
	RequestID  string
	IP         string
	Host       string
	Path       string
	Method     string
	UserAgent  string
	Env        string
	AppName    string
	AppVersion string
	CommitHash string
}
