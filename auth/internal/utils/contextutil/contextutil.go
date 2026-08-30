package contextutil

import (
	"context"

	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/entities"
)

func HasPermission(ctx context.Context, permission string) bool {
	permSet, ok := ctx.Value(constants.CtxPermissions).(map[string]any)
	if ok {
		_, found := permSet[permission]
		return found
	}
	return false
}

func GetAccount(ctx context.Context) *entities.Account {
	account, ok := ctx.Value(constants.CtxAccount).(*entities.Account)
	if ok && account != nil {
		return account
	}
	return nil
}

func GetAccessToken(ctx context.Context) *entities.AccessToken {
	token, ok := ctx.Value(constants.CtxAccessToken).(*entities.AccessToken)
	if ok && token != nil {
		return token
	}
	return nil
}

func GetFields(ctx context.Context, keys []string) map[string]any {
	m := make(map[string]any, len(keys))
	for _, k := range keys {
		if v := ctx.Value(k); v != nil {
			m[k] = v
		}
	}
	return m
}
