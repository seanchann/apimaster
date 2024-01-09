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

package authenticator

import (
	"fmt"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/seanchann/apimaster/pkg/auth"
	"github.com/seanchann/apimaster/pkg/auth/authenticator/internal/jwt"
	loginapi "github.com/seanchann/apimaster/pkg/auth/authenticator/internal/login"
	"github.com/xsbull/utils/logger"
)

// LoginAuth Login
type LoginAuth struct {
	jwtAuth        *jwt.JWTAuth
	loginApi       *loginapi.LoginApi
	authUserHandle auth.AuthenticationUser
}

// NewLoginUserManager  manager
func NewLoginAuth(jwtSecret []byte, expire time.Duration) auth.APIAuthenticator {

	manager := &LoginAuth{}
	manager.jwtAuth = jwt.NewJWTAuth(jwtSecret, expire)
	manager.loginApi = loginapi.NewLoginApi(manager)

	return manager
}

func (la *LoginAuth) GenerateAuthToken(username, namespace, uid string, groups []string) (token string, err error) {
	return la.jwtAuth.GenerateToken(username, namespace, uid, groups)
}

func (la *LoginAuth) InstallLoginAndJWTWebHook(ws *restful.WebService, authUserHandle auth.AuthenticationUser) {
	la.loginApi.Install(ws)
	la.jwtAuth.Install(ws)

	la.authUserHandle = authUserHandle
}

// LoginCheck 登录检测
func (la *LoginAuth) LoginCheck(username, namespace, password string) (token string, err error) {

	identityUser, err := la.authUserHandle.CheckUserInfo(username, namespace, password)
	if err != nil {
		logger.Logf(logger.ErrorLevel, "get user failed, err:%v", err)
		return "", fmt.Errorf("authentication failed(%v). user(%v) or password invalid", err, username)
	}

	return la.jwtAuth.GenerateToken(username, namespace, identityUser.UID, identityUser.Groups)
}

// Logout 登录退出
func (la *LoginAuth) Logout(username string, token string) (err error) {

	sess := la.jwtAuth.Validate(token)
	if sess == nil {
		return fmt.Errorf("invalid token with user=%v", username)
	}

	la.authUserHandle.LogoutUser(sess.Username, sess.Namespace, sess.UID, sess.Groups)

	return nil
}
