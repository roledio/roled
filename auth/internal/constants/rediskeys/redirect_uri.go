package rediskeys

import (
	"net/url"
)

func RedirectURIsByProjectID(projectID string) string {
	return "redirect_uris:project:" + projectID
}

func RedirectURIByProjectIDAndURI(projectID, redirectURI string) string {
	return "redirect_uri:project:" + projectID + ":uri:" + url.QueryEscape(redirectURI)
}
