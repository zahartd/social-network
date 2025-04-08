package auth

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const UserIDMetadataKey = "x-user-id"

type contextKey string

const userIDKey contextKey = "userID"

func AuthInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return handler(ctx, req)
	}

	userIDValues := md.Get(UserIDMetadataKey)
	var userID string
	if len(userIDValues) > 0 {
		userID = userIDValues[0]
	}

	if userID == "" {
		return handler(ctx, req)
	}

	newCtx := context.WithValue(ctx, userIDKey, userID)
	return handler(newCtx, req)
}

func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(userIDKey).(string)
	if !ok || userID == "" {
		return "", status.Errorf(codes.Unauthenticated, "user ID not found in context")
	}
	return userID, nil
}
