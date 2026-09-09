package rediskeys

import "fmt"

func UserIdentityByProviderAndProviderUserIDAndProjectID(
	provider,
	providerUserID,
	projectID string) string {
	return fmt.Sprintf("user_identity:provider:%s:provider_user_id:%s:project:%s",
		provider, providerUserID, projectID)
}

func UserIdentityByID(id string) string {
	return "user_identity:" + id
}
