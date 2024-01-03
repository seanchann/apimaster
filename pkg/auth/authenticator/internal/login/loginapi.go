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

	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/api/errors"
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
	Logout(name string) (err error)
}

type LoginApi struct {
	loginHook Interface
}

func NewLoginApi(handle Interface) *LoginApi {
	return &LoginApi{
		loginHook: handle,
	}
}

func (l *LoginApi) Install(ws *restful.WebService) {
	ws.Route(ws.POST("/apis/auth/logout").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		To(l.logout))
	ws.Route(ws.POST("/apis/auth/login").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		To(l.login))
}

func (l *LoginApi) login(req *restful.Request, resp *restful.Response) {

	loginSpec := &LoginSpec{}

	err := req.ReadEntity(loginSpec)
	if err != nil {
		apierr := errors.NewBadRequest(err.Error())
		resp.WriteErrorString(http.StatusBadRequest, apierr.Error())
		return
	}

	token, err := l.loginHook.LoginCheck(loginSpec.Username, loginSpec.Namespace, loginSpec.Password)
	if err != nil {
		apierr := errors.NewBadRequest(err.Error())
		resp.WriteErrorString(http.StatusBadRequest, apierr.Error())
	}

	loginSpec.Token = token
	loginSpec.Password = ""

	resp.WriteEntity(loginSpec)
}

func (l *LoginApi) logout(req *restful.Request, resp *restful.Response) {

	loginSpec := &LoginSpec{}

	err := req.ReadEntity(loginSpec)
	if err != nil {
		apierr := errors.NewBadRequest(err.Error())
		resp.WriteErrorString(http.StatusBadRequest, apierr.Error())
		return
	}

	resp.WriteEntity(loginSpec)
}
