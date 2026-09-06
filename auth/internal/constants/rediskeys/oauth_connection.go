package rediskeys

import (
	"fmt"
)

func OAuthConnectionsByProjectID(projectID string) string {
	return fmt.Sprintf("oauth_connection:project:%s", projectID)
}

func OAuthConnectionByProjectIDAndProvider(projectID, provider string) string {
	return fmt.Sprintf("oauth_connection:project:%s:provider:%s", projectID, provider)
}
