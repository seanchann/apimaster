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
)

type LoginSpec struct {
	Username  string `json:"username,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Password  string `json:"password,omitempty"`
	Token     string `json:"token,omitempty"`
}

// Login login interface
type Interface interface {
	//LoginCheck check user pwd
	LoginCheck(username, namespace, password string) (token string, err error)
	//Logout logout
	Logout(username string, token string) (err error)
}

type LoginApi struct {
	loginHook Interface
}

func NewLoginApi(handle Interface) *LoginApi {
	return &LoginApi{
		loginHook: handle,
	}
}

func (l *LoginApi) Login(req *restful.Request, resp *restful.Response) {

	loginSpec := &LoginSpec{}

	loginErr := NewLoginAuthError()

	err := req.ReadEntity(loginSpec)
	if err != nil {
		resp.WriteErrorString(http.StatusForbidden, loginErr.Status())
		return
	}

	token, err := l.loginHook.LoginCheck(loginSpec.Username, loginSpec.Namespace, loginSpec.Password)
	if err != nil {
		resp.WriteErrorString(http.StatusForbidden, loginErr.Status())
		return
	}

	loginSpec.Token = token
	loginSpec.Password = ""

	resp.WriteEntity(loginSpec)
}

func (l *LoginApi) Logout(req *restful.Request, resp *restful.Response) {

	loginSpec := &LoginSpec{}

	loginErr := NewLogoutError()

	err := req.ReadEntity(loginSpec)
	if err != nil {
		resp.WriteErrorString(http.StatusForbidden, loginErr.Status())
		return
	}

	token, found := strings.CutPrefix(req.Request.Header["Authorization"][0], "Bearer ")
	if !found {
		resp.WriteErrorString(http.StatusForbidden, loginErr.Status())
		return
	}

	if err = l.loginHook.Logout(loginSpec.Username, token); err != nil {
		resp.WriteErrorString(http.StatusForbidden, loginErr.Status())
		return
	}

	loginSpec.Token = ""
	loginSpec.Password = ""

	resp.WriteEntity(loginSpec)
}
