package rediskeys

func UserByID(userID string) string {
	return "user:" + userID
}

func UserByIDAndProjectID(userID, projectID string) string {
	return "user:" + userID + ":project:" + projectID
}

func UserByProjectIDAndEmail(projectID, email string) string {
	return "user:project:" + projectID + ":email:" + email
}

func UserByProjectIDAndExternalUserID(projectID, externalUserID string) string {
	return "user:project:" + projectID + ":external:" + externalUserID
}
