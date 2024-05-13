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
