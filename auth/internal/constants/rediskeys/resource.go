package rediskeys

func ResourceByIDAndProjectID(resourceID, projectID string) string {
	return "resource:" + resourceID + ":project:" + projectID
}

func ResourceByProjectIDAndCode(projectID, code string) string {
	return "resource:project:" + projectID + ":code:" + code
}
