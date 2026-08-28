package constants

const (
	AccessTokenStatusInit    = "init"    // Initial status when authorization code is created
	AccessTokenStatusIssued  = "issued"  // Status when access token is issued after exchanging authorization code, client credentials, or refresh token
	AccessTokenStatusExpired = "expired" // Status when access token is expired
	AccessTokenStatusRevoked = "revoked" // Status when access token is revoked
)
