package utils

import (
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInvalidUserID   = status.Error(codes.Internal, "internal error: invalid user ID format in context")
	ErrInvalidPage     = fmt.Errorf("page must be a positive integer")
	ErrInvalidPageSize = fmt.Errorf("page_size must be a positive integer")
	ErrInvalidPostID   = status.Error(codes.Internal, "internal error: invalid post ID format")
)

func ValidateUserID(userIDValue any) error {
	userID, ok := userIDValue.(string)
	if !ok || userID == "" {
		return ErrInvalidUserID
	}
	_, err := uuid.Parse(userID)
	if err != nil {
		return ErrInvalidUserID
	}
	return err
}

func ValidatePage(pageStr string) (int, error) {
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		return 0, ErrInvalidPage
	}
	return page, nil
}

func ValidatePageSize(pageSizeStr string) (int, error) {
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		return 0, ErrInvalidPage
	}
	return pageSize, nil
}

func ValidatePostID(postIDValue any) error {
	postID, ok := postIDValue.(string)
	if !ok || postID == "" {
		return ErrInvalidPostID
	}
	_, err := uuid.Parse(postID)
	if err != nil {
		return ErrInvalidPostID
	}
	return err
}
