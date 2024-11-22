package auth

import "context"

func GetToken(ctx context.Context) string {
	value, _ := ctx.Value(bearerKey).(string)
	return value
}

func GetUsername(ctx context.Context) string {
	value, _ := ctx.Value(usernameKey).(string)
	return value
}
