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

package apiserver

import (
	"time"

	"github.com/seanchann/apimaster/pkg/auth"
	"github.com/seanchann/apimaster/pkg/auth/authenticator"
	"github.com/seanchann/apimaster/pkg/auth/authorizer/rbac"
)

type APIAuthConfig struct {
	JWTAuthSecret      []byte
	JWTAuthexpire      time.Duration
	UserAuthentication auth.AuthenticationUser
	UserAuthorization  auth.AuthorizationUser
}

type apiAuth struct {
	auth.APIAuthenticator
	auth.APIAuthorizer
}

func NewAPIAuthHandle(conf APIAuthConfig) auth.Interface {
	impl := &apiAuth{}
	impl.APIAuthenticator = authenticator.NewLoginAuth(conf.JWTAuthSecret, conf.JWTAuthexpire, conf.UserAuthentication)
	impl.APIAuthorizer = rbac.NewAuthorizerRBAC(conf.UserAuthorization)

	return impl
}
