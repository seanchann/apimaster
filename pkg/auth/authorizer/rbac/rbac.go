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
	"github.com/xsbull/utils/logger"
	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AuthorizerRBAC struct {
}

// NewLoginUserManager  manager
func NewAuthorizerRBAC() auth.APIAuthorizer {
	rbac := &AuthorizerRBAC{}

	return rbac
}

func (ar *AuthorizerRBAC) InstallRBACWebHook(ws *restful.WebService) {
	ws.Route(ws.POST("/apis/auth/authorization").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		To(ar.RBACHandler))
}

func (ar *AuthorizerRBAC) RBACHandler(req *restful.Request, resp *restful.Response) {
	subjectReq := &authorizationv1.SubjectAccessReview{
		TypeMeta: metav1.TypeMeta{APIVersion: authorizationv1.SchemeGroupVersion.Version},
	}
	resp.Header().Set("Content-Type", "application/json")

	type status struct {
		Allowed         bool   `json:"allowed"`
		Reason          string `json:"reason"`
		EvaluationError string `json:"evaluationError"`
	}
	subjectResp := struct {
		APIVersion string `json:"apiVersion"`
		Kind       string `json:"kind"`
		Status     status `json:"status"`
	}{
		APIVersion: subjectReq.APIVersion, //"authorization.k8s.io/v1beta1", //authorizationv1.SchemeGroupVersion.String(),
		Kind:       "SubjectAccessReview",
		Status:     status{Allowed: false},
	}

	err := req.ReadEntity(subjectReq)
	if err != nil {
		logger.Logf(logger.ErrorLevel, "read SujectAccessReview body failed: %v", err)
		resp.WriteEntity(subjectResp)
		return
	}

	logger.Logf(logger.DebugLevel, "request SujectAccessReview body: %v", subjectReq)

	resp.WriteEntity(subjectResp)
	return
}
