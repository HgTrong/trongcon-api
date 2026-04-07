package jwtutil

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	UserID uint     `json:"uid"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

func Issue(userID uint, roles []string, secret []byte, exp time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Roles:  roles,
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
