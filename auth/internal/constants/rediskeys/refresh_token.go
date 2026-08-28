package rediskeys

func RefreshTokenByClientIDAndTokenHash(clientID, tokenHash string) string {
	return "refresh_token:client:" + clientID + ":hash:" + tokenHash
}
