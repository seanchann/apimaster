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

package authorizer

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/seanchann/apimaster/pkg/auth"
	"github.com/seanchann/apimaster/pkg/auth/authorizer/rbac"
)

type AuthorizerFactory struct {
	authRBAC *rbac.AuthorizerRBAC
}

func NewAuthorizer(permitHandle auth.AuthorizationHook) auth.APIAuthorizer {
	handle := &AuthorizerFactory{}

	handle.authRBAC = rbac.NewAuthorizerRBAC(permitHandle)

	return handle
}

func (af AuthorizerFactory) RBACHandler() restful.RouteFunction {
	return af.authRBAC.RBACWebHookHandler
}
