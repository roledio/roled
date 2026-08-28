package rediskeys

func ProjectSettingByProjectID(projectID string) string {
	return "project_setting:project:" + projectID
}
