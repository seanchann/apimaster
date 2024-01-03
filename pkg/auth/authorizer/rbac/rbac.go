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

package rbac

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/seanchann/apimaster/pkg/auth"
)

type AuthorizerRBAC struct {
}

// NewLoginUserManager  manager
func NewAuthorizerRBAC() auth.APIAuthorizer {
	rbac := &AuthorizerRBAC{}

	return rbac
}

func (ar *AuthorizerRBAC) InstallRBACWebHook(ws *restful.WebService) {

}
