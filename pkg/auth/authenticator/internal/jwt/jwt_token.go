/********************************************************************
* Copyright (c) 2008 - 2024. sean <seandev@foxmail>
* All rights reserved.
*
* PROPRIETARY RIGHTS of the following material in either
* electronic or paper format pertain to sean.
* All manufacturing, reproduction, use, and sales involved with
* this subject MUST conform to the license agreement signed
* with sean.
*******************************************************************/

package jwt

import (
	"fmt"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/xsbull/utils/logger"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var hmacSampleSecret = []byte("41c5b3a9afc98ceb366d5c9db81291")

// SessionInfo session info store some information for request user
type SessionInfo struct {
	Username  string   `json:"username"`
	Groups    []string `json:"groups"`
	UID       string   `json:"uid"`
	Namespace string   `json:"namespace"`
}

type JWTClaims struct {
	SessionInfo
	jwt.RegisteredClaims
}

// ApiJWTToken store a string for unique client
type JWTAuth struct {
}

// NewJWTAuth new a api session
func NewJWTAuth() *JWTAuth {
	jwt := &JWTAuth{}

	return jwt
}

func (ja *JWTAuth) Install(ws *restful.WebService) {
	ws.Route(ws.POST("/apis/auth/authentication").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		To(ja.Authenticate))
}

// GenerateSession generate new session for user
func (ja *JWTAuth) GenerateToken(username string, namespace string, groups []string) (token string, err error) {
	claims := JWTClaims{
		SessionInfo{
			Username:  username,
			Groups:    groups,
			Namespace: namespace,
		},
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return jwtToken.SignedString(hmacSampleSecret)
}

func (ja *JWTAuth) Validate(token string) *SessionInfo {
	jwtToken, err := jwt.ParseWithClaims(token, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return hmacSampleSecret, nil
	})

	if err != nil {
		logger.Logf(logger.ErrorLevel, "validate token(%s) failed(%v)", token, err)
		return nil
	}

	if claims, ok := jwtToken.Claims.(*JWTClaims); ok && jwtToken.Valid {
		return &SessionInfo{
			Username:  claims.Username,
			Groups:    claims.Groups,
			Namespace: claims.Namespace,
		}
	}

	return nil
}

func (ja *JWTAuth) Authenticate(req *restful.Request, resp *restful.Response) {
	tokenReq := &authenticationv1.TokenReview{
		TypeMeta: metav1.TypeMeta{APIVersion: authenticationv1.SchemeGroupVersion.String()},
	}
	status := &authenticationv1.TokenReviewStatus{}

	tokenResp := struct {
		APIVersion string                              `json:"apiVersion"`
		Kind       string                              `json:"kind"`
		Status     *authenticationv1.TokenReviewStatus `json:"status"`
	}{
		APIVersion: tokenReq.APIVersion, // authenticationv1.SchemeGroupVersion.String(),
		Kind:       "TokenReview",
		Status:     status,
	}

	err := req.ReadEntity(tokenReq)
	if err != nil {
		logger.Logf(logger.ErrorLevel, "read TokenReview body failed: %v", err)
		resp.WriteEntity(tokenResp)
		return
	}

	logger.Logf(logger.DebugLevel, "authenticating TokenReview body: %v", tokenReq)

	if sessInfo := ja.Validate(tokenReq.Spec.Token); sessInfo != nil {

		tokenResp.Status = &authenticationv1.TokenReviewStatus{
			Authenticated: true,
			User: authenticationv1.UserInfo{
				Username: sessInfo.Username,
				UID:      sessInfo.UID,
				Groups:   append([]string{}, sessInfo.Groups...),
				Extra:    map[string]authenticationv1.ExtraValue{"namespace": {sessInfo.Namespace}},
			},
		}
	} else {
		logger.Logf(logger.ErrorLevel, "validating TokenRequest signature  %v", err)
	}

	resp.WriteEntity(tokenResp)
}