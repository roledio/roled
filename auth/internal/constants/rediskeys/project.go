package rediskeys

import (
	"fmt"
)

func ProjectByID(projectID string) string {
	return "project:" + projectID
}

func ProjectByIDAndAccountID(projectID, accountID string) string {
	return "project:" + projectID + ":account:" + accountID
}

func ProjectByIsSystem(isSystem bool) string {
	return "project:system:" + fmt.Sprintf("%v", isSystem)
}
