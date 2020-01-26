package application

import (
	"context"
	"regexp"
)

type User struct {
	Username string
}

type key int

const (
	userKey key = iota
)

var usernameRegexp = regexp.MustCompile("^[a-zA-Z0-9_][a-zA-Z0-9@_.\\-]*$")

func NewContext(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func UserFromContext(ctx context.Context) (*User, bool) {
	u, ok := ctx.Value(userKey).(*User)
	return u, ok
}

func ValidUsername(str string) bool {
	return usernameRegexp.MatchString(str)
}
