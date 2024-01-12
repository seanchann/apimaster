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

	token, found := strings.CutPrefix(req.Request.Header["Authorization"][0], "Bearer ")
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
