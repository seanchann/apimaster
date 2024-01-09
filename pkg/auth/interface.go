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

package auth

import (
	"github.com/emicklei/go-restful/v3"
	"k8s.io/apiserver/pkg/authentication/user"
)

type AuthenticationUser interface {
	CheckUserInfo(username, namespace, password string) (*user.DefaultInfo, error)
	LogoutUser(username, namespace, uid string, groups []string)
}

type APIAuthenticator interface {
	InstallLoginAndJWTWebHook(ws *restful.WebService, authUserHandle AuthenticationUser)
	GenerateAuthToken(username, namespace, uid string, groups []string) (token string, err error)
}

type AuthorizationNonResourceAttributes struct {
	// Path is the URL path of the request
	Path string
	// Verb is the standard HTTP verb
	Verb string
}

type AuthorizationResourceAttributes struct {
	// Namespace is the namespace of the action being requested.  Currently, there is no distinction between no namespace and all namespaces
	// "" (empty) is defaulted for LocalSubjectAccessReviews
	// "" (empty) is empty for cluster-scoped resources
	// "" (empty) means "all" for namespace scoped resources from a SubjectAccessReview or SelfSubjectAccessReview
	Namespace string
	// Verb is a kubernetes resource API verb, like: get, list, watch, create, update, delete, proxy.  "*" means all.
	Verb string
	// Group is the API Group of the Resource.  "*" means all.
	Group string
	// Version is the API Version of the Resource.  "*" means all.
	Version string
	// Resource is one of the existing resource types.  "*" means all.
	Resource string
	// Subresource is one of the existing resource types.  "" means none.
	Subresource string
	// Name is the name of the resource being requested for a "get" or deleted for a "delete". "" (empty) means all.
	Name string
}

type AuthorizationPermissions struct {
	Username      string
	UserGroup     []string
	UserUID       string
	UserExtraData map[string][]string

	NonResourceAttributes *AuthorizationNonResourceAttributes
	ResourceAttributes    *AuthorizationResourceAttributes
}

type AuthorizationUser interface {
	CheckUserPermissions(perm AuthorizationPermissions) bool
}

type APIAuthorizer interface {
	InstallRBACWebHook(ws *restful.WebService, permitHandle AuthorizationUser)
}

type Interface interface {
	APIAuthenticator
	APIAuthorizer
}
