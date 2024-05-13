/********************************************************************
* Copyright (c) 2008 - 2024. Authors: seanchann <seandev@foxmail.com>
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*         http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*******************************************************************/

package jwt

import (
	"fmt"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/golang-jwt/jwt/v5"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

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
	secret []byte
	expire time.Duration
}

// NewJWTAuth new a api session
func NewJWTAuth(secret []byte, expire time.Duration) *JWTAuth {
	jwt := &JWTAuth{
		secret: append([]byte{}, secret...),
		expire: expire,
	}

	return jwt
}

// GenerateDebugToken generate new session for user
func (ja *JWTAuth) GenerateDebugToken(username string, namespace, uid string, groups []string) (token string, err error) {
	claims := JWTClaims{
		SessionInfo{
			Username:  username,
			Groups:    groups,
			Namespace: namespace,
			UID:       uid,
		},
		jwt.RegisteredClaims{
			ID: uid,
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return jwtToken.SignedString(ja.secret)
}

// GenerateSession generate new session for user
func (ja *JWTAuth) GenerateToken(username string, namespace, uid string, groups []string, timeout time.Duration) (token string, err error) {
	claims := JWTClaims{
		SessionInfo{
			Username:  username,
			Groups:    groups,
			Namespace: namespace,
			UID:       uid,
		},
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(timeout)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return jwtToken.SignedString(ja.secret)
}

func (ja *JWTAuth) Validate(token string) *SessionInfo {
	jwtToken, err := jwt.ParseWithClaims(token, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return ja.secret, nil
	})

	if err != nil {
		klog.Errorf("validate token(%s) failed(%v)", token, err)
		return nil
	}

	if claims, ok := jwtToken.Claims.(*JWTClaims); ok && jwtToken.Valid {
		return &SessionInfo{
			Username:  claims.Username,
			Groups:    claims.Groups,
			Namespace: claims.Namespace,
			UID:       claims.UID,
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
		klog.Errorf("read TokenReview body failed: %v", err)
		resp.WriteEntity(tokenResp)
		return
	}

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
		klog.Errorf("validating TokenRequest signature  %v", err)
	}

	resp.WriteEntity(tokenResp)
}
