package rediskeys

import (
	"fmt"
)

func PermissionsByRoleID(roleID string) string {
	return "permissions:role:" + roleID
}

func PermissionsByClientID(clientID string) string {
	return "permissions:client:" + clientID
}

func PermissionsByProjectIDAndIsDefault(projectID string, isDefault *bool) string {
	return "permissions:project:" + projectID + ":default:" + fmt.Sprintf("%v", isDefault)
}
