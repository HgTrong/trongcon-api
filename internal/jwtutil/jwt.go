package jwtutil

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

const (
	PurposeAccess        = "access"
	PurposePasswordReset = "password_reset"
	PurposeGymCheckIn    = "gym_checkin"
)

type Claims struct {
	UserID       uint     `json:"uid"`
	Roles        []string `json:"roles"`
	Purpose      string   `json:"purpose,omitempty"`
	MembershipID uint     `json:"mid,omitempty"`
	jwt.RegisteredClaims
}

func Issue(userID uint, roles []string, secret []byte, exp time.Duration) (string, error) {
	return IssueWithPurpose(userID, roles, PurposeAccess, secret, exp)
}

func IssueWithPurpose(userID uint, roles []string, purpose string, secret []byte, exp time.Duration) (string, error) {
	return IssueCheckIn(userID, 0, roles, purpose, secret, exp)
}

func IssueCheckIn(userID, membershipID uint, roles []string, purpose string, secret []byte, exp time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:       userID,
		Roles:        roles,
		Purpose:      purpose,
		MembershipID: membershipID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(exp)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(secret)
}

func Parse(tokenString string, secret []byte) (*Claims, error) {
	t, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}
	claims, ok := t.Claims.(*Claims)
	if !ok || !t.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func ParseWithPurpose(tokenString string, purpose string, secret []byte) (*Claims, error) {
	claims, err := Parse(tokenString, secret)
	if err != nil {
		return nil, err
	}
	got := claims.Purpose
	if got == "" {
		got = PurposeAccess
	}
	if got != purpose {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
