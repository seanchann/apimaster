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

package login

import (
	"net/http"
	"strings"

	"github.com/emicklei/go-restful/v3"
	"github.com/seanchann/apimaster/pkg/auth"
)

type LoginApi struct {
	loginHook auth.AuthenticationHook
}

func NewLoginApi(handle auth.AuthenticationHook) *LoginApi {
	return &LoginApi{
		loginHook: handle,
	}
}

func (l *LoginApi) Login(req *restful.Request, resp *restful.Response) {

	loginErr := NewLoginAuthError()

	loginCheck := func(readObj interface{}) error {
		return req.ReadEntity(readObj)
	}

	respBody, err := l.loginHook.Login(loginCheck)
	if err != nil {
		resp.WriteErrorString(http.StatusForbidden, loginErr.Status())
		return
	}

	resp.WriteEntity(respBody)
}

func (l *LoginApi) Logout(req *restful.Request, resp *restful.Response) {
	loginErr := NewLogoutError()

	authHeader := req.Request.Header["Authorization"]
	if len(authHeader) == 0 {
		resp.WriteErrorString(http.StatusForbidden, loginErr.Status())
		return
	}
	token, found := strings.CutPrefix(authHeader[0], "Bearer ")
	if !found {
		resp.WriteErrorString(http.StatusForbidden, loginErr.Status())
		return
	}

	logoutCheck := func(readObj interface{}) error {
		return req.ReadEntity(readObj)
	}

	respBody, err := l.loginHook.Logout(logoutCheck, token)
	if err != nil {
		resp.WriteErrorString(http.StatusForbidden, loginErr.Status())
		return
	}

	resp.WriteEntity(respBody)
}
