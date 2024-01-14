/********************************************************************
* Copyright (c) 2008 - 2024. seanchann <seanchann.zhou@gmail.com>
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
	"time"

	"github.com/emicklei/go-restful/v3"
)

// user.DefaultInfo扩展的Extra字段的key
var UserDefaultInfoExtraKeyNamespace = "namespace"

type LoginCheckFunc func(readObj interface{}) error

type AuthenticationHook interface {
	Login(checkFunc LoginCheckFunc) (resp interface{}, err error)
	Logout(checkFunc LoginCheckFunc, token string) (respBody interface{}, err error)
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

type UserInfo struct {
	Username      string
	UserGroup     []string
	UserUID       string
	UserExtraData map[string][]string
}

type AuthorizationPermissions struct {
	UserInfo
	NonResourceAttributes *AuthorizationNonResourceAttributes
	ResourceAttributes    *AuthorizationResourceAttributes
}

type AuthorizationHook interface {
	CheckUserPermissions(perm AuthorizationPermissions) bool
}

type APIAuthorizer interface {
	//JWTTokenHandler 安装支持类k8s的auth webhook的处理
	RBACHandler() restful.RouteFunction
}

type APIAuthenticator interface {
	LoginHandler() restful.RouteFunction
	LogoutHandler() restful.RouteFunction
	//JWTTokenHandler 安装支持类k8s的auth webhook的处理
	JWTTokenHandler() restful.RouteFunction

	GenerateAuthToken(username, namespace, uid string, groups []string, timeout time.Duration) (token string, err error)
	Validate(token string) (*UserInfo, error)
}

type Interface interface {
	APIAuthenticator
	APIAuthorizer
}
