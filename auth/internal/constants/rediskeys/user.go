package rediskeys

func UserByID(userID string) string {
	return "user:" + userID
}

func UserByProjectIDAndEmail(projectID, email string) string {
	return "user:project:" + projectID + ":email:" + email
}

func UserByProjectIDAndExternalUserID(projectID, externalUserID string) string {
	return "user:project:" + projectID + ":external:" + externalUserID
}
