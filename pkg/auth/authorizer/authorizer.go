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
