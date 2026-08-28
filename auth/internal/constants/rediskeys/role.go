package rediskeys

func RoleByIDAndProjectID(roleID, projectID string) string {
	return "role:" + roleID + ":project:" + projectID
}

func RoleByProjectIDAndCode(projectID, code string) string {
	return "role:project:" + projectID + ":code:" + code
}

func RoleByUserID(userID string) string {
	return "role:user:" + userID
}
