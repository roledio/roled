package rediskeys

func UserIdentityByProviderAndProviderUserID(provider, providerUserID string) string {
	return "user_identity:provider:" + provider + ":provider_user_id:" + providerUserID
}

func UserIdentityByID(id string) string {
	return "user_identity:" + id
}
