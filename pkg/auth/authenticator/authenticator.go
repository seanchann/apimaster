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
func NewLoginAuth() auth.APIAuthenticator {

	manager := &LoginAuth{}
	manager.jwtAuth = jwt.NewJWTAuth()
	manager.loginApi = loginapi.NewLoginApi(manager)

	return manager
}

func (la LoginAuth) InstallLoginAndJWTWebHook(ws *restful.WebService, authUserHandle auth.AuthenticationUser) {
	la.loginApi.Install(ws)
	la.jwtAuth.Install(ws)

	la.authUserHandle = authUserHandle
}

// LoginCheck 登录检测
func (la *LoginAuth) LoginCheck(username, namespace, password string) (token string, err error) {

	identityUser, err := la.authUserHandle.GetUserInfo(username, namespace, password)
	if err != nil {
		logger.Logf(logger.ErrorLevel, "get user failed, err:%v", err)
		return "", fmt.Errorf("Authentication failed. user or password invalid")
	}

	return la.jwtAuth.GenerateToken(username, namespace, identityUser.Groups)
}

// Logout 登录退出
func (la *LoginAuth) Logout(username string) error {
	// obj, found := loginu.user2sid.Get(username)
	// if !found {
	// 	logger.Log(logger.ErrorLevel, "logout user=%v not found", username)
	// 	return fmt.Errorf("logout user=%v not found", username)
	// }

	// sid, ok := obj.(string)
	// if !ok {
	// 	logger.Log(logger.ErrorLevel, "logout user=%v not found sid", username)
	// 	return fmt.Errorf("logout user=%v not found sid", username)
	// }

	// loginu.Delete(sid)
	return nil
}
