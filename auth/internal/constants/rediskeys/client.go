package rediskeys

import (
	"fmt"
)

func ClientByID(clientID string) string {
	return "client:" + clientID
}

func ClientByIDAndProjectID(clientID, projectID string) string {
	return "client:" + clientID + ":project:" + projectID
}

func ClientByProjectIDAndIsDefault(projectID string, isDefault bool) string {
	return "client:project:" + projectID + ":default:" + fmt.Sprintf("%v", isDefault)
}
