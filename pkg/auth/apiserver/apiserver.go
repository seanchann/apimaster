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
	"github.com/seanchann/apimaster/pkg/auth"
	"github.com/seanchann/apimaster/pkg/auth/authenticator"
	"github.com/seanchann/apimaster/pkg/auth/authorizer/rbac"
)

type apiAuth struct {
	auth.APIAuthenticator
	auth.APIAuthorizer
}

func NewAPIAuthHandle() auth.Interface {
	impl := &apiAuth{}
	impl.APIAuthenticator = authenticator.NewLoginAuth()
	impl.APIAuthorizer = rbac.NewAuthorizerRBAC()

	return impl
}
