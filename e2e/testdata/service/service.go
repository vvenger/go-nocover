package service

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
)

var ErrForbidden = errors.New("forbidden")

type URLService struct {
	blackListUsers map[int]struct{}
}

func New() *URLService {
	return &URLService{
		blackListUsers: map[int]struct{}{
			2: {},
		},
	}
}

func (s *URLService) isBlocked(userID int) bool {
	_, ok := s.blackListUsers[userID]
	return ok
}

func (s *URLService) Encode(userID int, input string) (string, error) {
	if s.isBlocked(userID) {
		slog.Debug("debug log", "userID", userID)
		return "", ErrForbidden
	}

	slog.Debug("debug log",
		"userID",
		userID)

	return url.QueryEscape(input), nil
}

func (s *URLService) Decode(userID int, input string) (string, error) {
	if s.isBlocked(userID) {
		return "", ErrForbidden
	}

	decoded, err := url.QueryUnescape(input)

	if err != nil {
		return "", fmt.Errorf("decode error: %w", err)
	}
	return decoded, nil
}
