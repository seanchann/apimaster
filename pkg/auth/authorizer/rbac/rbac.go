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

package rbac

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/seanchann/apimaster/pkg/auth"
	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

type AuthorizerRBAC struct {
	authHandle auth.AuthorizationHook
}

// NewLoginUserManager  manager
func NewAuthorizerRBAC(permitHandle auth.AuthorizationHook) *AuthorizerRBAC {
	rbac := &AuthorizerRBAC{
		authHandle: permitHandle,
	}

	return rbac
}

func (ar *AuthorizerRBAC) RBACWebHookHandler(req *restful.Request, resp *restful.Response) {

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
		APIVersion: subjectReq.APIVersion,
		Kind:       "SubjectAccessReview",
		Status:     status{Allowed: false},
	}

	err := req.ReadEntity(subjectReq)
	if err != nil {
		klog.Infof("read SujectAccessReview body failed: %v", err)
		resp.WriteEntity(subjectResp)
		return
	}

	permission := auth.AuthorizationPermissions{
		UserInfo: auth.UserInfo{
			Username:      subjectReq.Spec.User,
			UserUID:       subjectReq.Spec.UID,
			UserGroup:     subjectReq.Spec.Groups,
			UserExtraData: make(map[string][]string),
		},
	}

	if subjectReq.Spec.NonResourceAttributes != nil {
		permission.NonResourceAttributes = &auth.AuthorizationNonResourceAttributes{}

		permission.NonResourceAttributes.Path = subjectReq.Spec.NonResourceAttributes.Path
		permission.NonResourceAttributes.Verb = subjectReq.Spec.NonResourceAttributes.Verb
	}

	if subjectReq.Spec.ResourceAttributes != nil {
		permission.ResourceAttributes = &auth.AuthorizationResourceAttributes{}

		permission.ResourceAttributes.Group = subjectReq.Spec.ResourceAttributes.Group
		permission.ResourceAttributes.Name = subjectReq.Spec.ResourceAttributes.Name
		permission.ResourceAttributes.Namespace = subjectReq.Spec.ResourceAttributes.Namespace
		permission.ResourceAttributes.Resource = subjectReq.Spec.ResourceAttributes.Resource
		permission.ResourceAttributes.Subresource = subjectReq.Spec.ResourceAttributes.Subresource
		permission.ResourceAttributes.Verb = subjectReq.Spec.ResourceAttributes.Verb
		permission.ResourceAttributes.Version = subjectReq.Spec.ResourceAttributes.Version
	}

	for key, v := range subjectReq.Spec.Extra {
		permission.UserExtraData[key] = append([]string{}, v...)
	}

	subjectResp.Status.Allowed = ar.authHandle.CheckUserPermissions(permission)

	klog.V(3).Infof("request SujectAccessReview body: %v", subjectResp)

	resp.WriteEntity(subjectResp)

}
