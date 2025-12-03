package jwt

import (
	"context"
	"fmt"
	"strconv"

	"google.golang.org/grpc/metadata"
)

type UserContext struct {
	context.Context
	Userid int64
	Token  string
}

func WithValues(ctx context.Context, userid int64, token string) *UserContext {
	return &UserContext{
		Context: ctx,
		Userid:  userid,
		Token:   token,
	}
}

func FromContext(ctx context.Context) (*UserContext, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to get metadata from incoming context")
	}

	token := md.Get("token")
	userid := md.Get("userid")

	if len(token) == 0 || len(userid) == 0 {
		return nil, fmt.Errorf("did not find the neccessary user info in context (userid: %v, token: %v)", userid, token)
	}

	userIdN, err := strconv.Atoi(userid[0])
	if err != nil {
		return nil, fmt.Errorf("failed to convert userid to integer: %v", err)
	}

	return &UserContext{
		Context: context.Background(),
		Userid:  int64(userIdN),
		Token:   token[0],
	}, nil
}
