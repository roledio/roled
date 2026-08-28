package rediskeys

func AccountByID(accountID string) string {
	return "account:" + accountID
}
