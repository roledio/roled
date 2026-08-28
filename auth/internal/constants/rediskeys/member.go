package rediskeys

func MemberByID(memberID string) string {
	return "member:" + memberID
}

func MemberByIDJoin(memberID string) string {
	return "member_join:" + memberID
}

func MemberByAccountIDAndUserID(accountID, userID string) string {
	return "member:account:" + accountID + ":user:" + userID
}
