package rediskeys

func AuthCodeByClientIDAndCodeHash(clientID, codeHash string) string {
	return "auth_code:client:" + clientID + ":hash:" + codeHash
}
