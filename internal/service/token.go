package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"gin-demo/internal/model"
	"gin-demo/internal/repository"
	"gin-demo/pkg/config"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	AccessTokenTTL  = 24 * time.Hour
	RefreshTokenTTL = 7 * 24 * time.Hour
)

func createAccessToken(user *model.User, sid string) (string, error) {
	jtiBytes, err := randomBytes(16)
	if err != nil {
		return "", err
	}
	jti := hex.EncodeToString(jtiBytes)
	claims := jwt.MapClaims{
		"exp":      time.Now().Add(AccessTokenTTL).Unix(),
		"sub":      strconv.FormatUint(uint64(user.ID), 10),
		"role":     user.Role,
		"sid":      sid,
		"jti":      jti,
		"type":     "access",
		"iat":      time.Now().Unix(),
		"username": user.Username,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.JWTSecret))

}

func ParseToken(tokenStr string) (*model.TokenInfo, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		return []byte(config.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	claims := token.Claims.(jwt.MapClaims)
	rawUserID, ok1 := claims["sub"].(string)
	if !ok1 {
		return nil, ErrInvalidToken
	}
	userID1, err := strconv.ParseUint(rawUserID, 10, 64)
	if err != nil {
		return nil, err
	}
	userID := uint(userID1)
	role, ok := claims["role"].(string)
	username, _ := claims["username"].(string)
	type1, _ := claims["type"].(string)
	sid, _ := claims["sid"].(string)
	if !ok || type1 != "access" || sid == "" {
		return nil, ErrInvalidToken

	}
	return &model.TokenInfo{
		UserID:   userID,
		Username: username,
		Role:     role,
		SID:      sid,
	}, nil
}
func randomBytes(n int) ([]byte, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func sha256Hex(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func GenerateTokenPair(
	user *model.User,
	platform string,
	ip string,
	userAgent string,
) (accessToken string, refreshToken string, err error) {
	sidebyte, err := randomBytes(16)
	if err != nil {
		return "", "", err
	}
	sid := hex.EncodeToString(sidebyte)
	access, err := createAccessToken(user, sid)
	if err != nil {
		return "", "", err
	}
	refreshbyte, err := randomBytes(32)
	if err != nil {
		return "", "", err
	}
	refresh := base64.RawURLEncoding.EncodeToString(refreshbyte)

	s := &model.Session{
		ID:               sid,
		UserID:           user.ID,
		AccessTokenHash:  sha256Hex(access),
		RefreshTokenHash: sha256Hex(refresh),
		PrevRefreshHash:  "",
		AccessExpiresAt:  time.Now().Add(AccessTokenTTL),
		RefreshExpiresAt: time.Now().Add(RefreshTokenTTL),
		Platform:         platform,
		IP:               ip,
		UserAgent:        userAgent,
		CreatedAt:        time.Now(),
		LastUsedAt:       time.Now(),
	}
	err = repository.CreateSession(s)
	if err != nil {
		return "", "", err
	}

	return access, refresh, nil

}
func ValidateToken(tokenStr string) (*model.TokenInfo, error) {
	info, err := ParseToken(tokenStr)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}
	hash := sha256Hex(tokenStr)
	s, err := repository.FindByAccessHash(hash)
	if err != nil {
		return nil, ErrSessionInvalid
	}
	if s.RevokedAt != nil {
		return nil, ErrSessionInvalid
	}

	return info, nil
}

func SessionCancel(sid string) error {
	return repository.RevokeSession(sid)

}

func RefreshTokens(refreshToken string) (access, refresh string, err error) {
	hash := sha256Hex(refreshToken)
	s, err := repository.FindByRefreshHash(hash)
	if err != nil {
		s1, err1 := repository.FindByPrevRefreshHash(hash)
		if err1 != nil {
			return "", "", ErrRefreshTokenInvalid
		}
		repository.RevokeSession(s1.ID)
		fmt.Println("❌刷新令牌已被使用，撤销会话:", s1.ID)
		return "", "", ErrRefreshTokenInvalid

	}
	if s.RevokedAt != nil || time.Now().After(s.RefreshExpiresAt) {
		return "", "", ErrRefreshTokenInvalid
	}
	refreshTokenHash, err := randomBytes(32)
	if err != nil {
		return "", "", err
	}
	refresh = base64.RawURLEncoding.EncodeToString(refreshTokenHash)
	user, err := repository.FindUserByID(s.UserID)
	if err != nil {
		return "", "", err
	}
	access, err = createAccessToken(user, s.ID)
	if err != nil {
		return "", "", err
	}
	err = repository.RotateTokens(s.ID, hash, sha256Hex(access), sha256Hex(refresh))
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
}
