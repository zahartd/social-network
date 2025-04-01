package auth

import (
	"context"
	"log"

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
		log.Printf("Auth Interceptor Warning: Missing metadata in incoming context for method %s", info.FullMethod)
		return handler(ctx, req)
	}

	userIDValues := md.Get(UserIDMetadataKey)
	var userID string
	if len(userIDValues) > 0 {
		userID = userIDValues[0]
	}

	if userID == "" {
		log.Printf("Auth Interceptor Warning: Missing '%s' in metadata for method %s", UserIDMetadataKey, info.FullMethod)
		return handler(ctx, req)
	}

	newCtx := context.WithValue(ctx, userIDKey, userID)
	log.Printf("Auth Interceptor: UserID '%s' extracted for method %s", userID, info.FullMethod)

	return handler(newCtx, req)
}

func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(userIDKey).(string)
	if !ok || userID == "" {
		log.Println("Error: User ID not found or invalid in context")
		return "", status.Errorf(codes.Unauthenticated, "user ID not found in context")
	}
	return userID, nil
}
