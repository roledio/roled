package rediskeys

func GoogleOAuthTransaction(state string) string {
	return "google_oauth:transaction:" + state
}
