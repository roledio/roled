package rediskeys

func AccessTokenByID(tokenID string) string {
	return "access_token:" + tokenID
}

func AccessTokenByIDJoin(tokenID string) string {
	return "access_token_join:" + tokenID
}
