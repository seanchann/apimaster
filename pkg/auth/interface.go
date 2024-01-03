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

package auth

import (
	"github.com/emicklei/go-restful/v3"
	"k8s.io/apiserver/pkg/authentication/user"
)

type AuthenticationUser interface {
	GetUserInfo(username, namespace, password string) (*user.DefaultInfo, error)
}

type APIAuthenticator interface {
	InstallLoginAndJWTWebHook(ws *restful.WebService, authUserHandle AuthenticationUser)
}

type APIAuthorizer interface {
	InstallRBACWebHook(ws *restful.WebService)
}

type Interface interface {
	APIAuthenticator
	APIAuthorizer
}
