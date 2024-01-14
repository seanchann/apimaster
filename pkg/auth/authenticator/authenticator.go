/********************************************************************
* Copyright (c) 2008 - 2024. seanchann <seanchann.zhou@gmail.com>
* All rights reserved.
*
* PROPRIETARY RIGHTS of the following material in either
* electronic or paper format pertain to sean.
* All manufacturing, reproduction, use, and sales involved with
* this subject MUST conform to the license agreement signed
* with sean.
*******************************************************************/

package authenticator

import (
	"fmt"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/seanchann/apimaster/pkg/auth"
	"github.com/seanchann/apimaster/pkg/auth/authenticator/internal/jwt"
	loginapi "github.com/seanchann/apimaster/pkg/auth/authenticator/internal/login"
)

// LoginAuth Login
type LoginAuth struct {
	jwtAuth        *jwt.JWTAuth
	loginApi       *loginapi.LoginApi
	authUserHandle auth.AuthenticationHook
}

// NewLoginUserManager  manager
func NewLoginAuth(jwtSecret []byte, expire time.Duration, authUserHandle auth.AuthenticationHook) auth.APIAuthenticator {

	manager := &LoginAuth{
		authUserHandle: authUserHandle,
	}
	manager.jwtAuth = jwt.NewJWTAuth(jwtSecret, expire)
	manager.loginApi = loginapi.NewLoginApi(manager.authUserHandle)

	return manager
}

func (la *LoginAuth) GenerateAuthToken(username, namespace, uid string, groups []string, timeout time.Duration) (token string, err error) {

	if timeout == 0 {
		return la.jwtAuth.GenerateDebugToken(username, namespace, uid, groups)
	}
	return la.jwtAuth.GenerateToken(username, namespace, uid, groups, timeout)
}

func (la *LoginAuth) LoginHandler() restful.RouteFunction {
	return la.loginApi.Login
}

func (la *LoginAuth) LogoutHandler() restful.RouteFunction {
	return la.loginApi.Logout
}

func (la *LoginAuth) JWTTokenHandler() restful.RouteFunction {
	return la.jwtAuth.Authenticate
}

func (la *LoginAuth) Validate(token string) (*auth.UserInfo, error) {
	sess := la.jwtAuth.Validate(token)
	if sess == nil {
		return nil, fmt.Errorf("invalid token with user=%v", sess.Username)
	}

	user := &auth.UserInfo{
		Username:  sess.Username,
		UserGroup: append([]string{}, sess.Groups...),
		UserUID:   sess.UID,
		UserExtraData: map[string][]string{
			auth.UserDefaultInfoExtraKeyNamespace: {sess.Namespace},
		},
	}

	return user, nil

}
