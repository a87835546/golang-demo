package helper

import (
	"github.com/dgrijalva/jwt-go"
)

type Claims struct {
	AdminName string `json:"admin_name"`
	jwt.StandardClaims
}

func GenerateToken(claims Claims, jwtsalt []byte) (string, error) {
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(jwtsalt)

	return token, err
}

func ParseToken(token string, jwtsalt string) (*Claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) { //返回一个token类型指针
		return []byte(jwtsalt), nil
	})
	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}

	return nil, err
}
