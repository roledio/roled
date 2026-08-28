package constants

const (
	RequestLogRequestID  = "request_id"
	RequestLogPath       = "path"
	RequestLogMethod     = "method"
	RequestLogIP         = "ip"
	RequestLogHost       = "host"
	RequestLogOrigin     = "origin"
	RequestLogReferer    = "referer"
	RequestLogUserAgent  = "user_agent"
	RequestLogEnv        = "env"
	RequestLogAppName    = "app_name"
	RequestLogAppVersion = "app_version"
	RequestLogCommitHash = "commit_hash"
)

var RequestLoggerKeys = []string{
	RequestLogRequestID,
	RequestLogPath,
	RequestLogMethod,
	RequestLogIP,
	RequestLogHost,
	RequestLogOrigin,
	RequestLogReferer,
	RequestLogUserAgent,
	RequestLogEnv,
	RequestLogAppName,
	RequestLogAppVersion,
	RequestLogCommitHash,
}
