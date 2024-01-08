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
	CheckUserInfo(username, namespace, password string) (*user.DefaultInfo, error)
	LogoutUser(username, namespace, uid string, groups []string)
}

type APIAuthenticator interface {
	InstallLoginAndJWTWebHook(ws *restful.WebService, authUserHandle AuthenticationUser)
}

type AuthorizationPermissions struct {
	Username      string
	UserGroup     []string
	UserUID       string
	Extra         map[string][]string
	APIKind       string
	APINamespace  string
	APIGroup      string
	RequestMethod string
}

type AuthorizationUser interface {
	CheckUserPermissions(perm AuthorizationPermissions) bool
}

type APIAuthorizer interface {
	InstallRBACWebHook(ws *restful.WebService, permitHandle AuthorizationUser)
}

type Interface interface {
	APIAuthenticator
	APIAuthorizer
}
