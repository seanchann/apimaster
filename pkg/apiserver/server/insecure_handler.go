/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"net/http"

	genericapifilters "k8s.io/apiserver/pkg/endpoints/filters"
	"k8s.io/apiserver/pkg/server"
	genericfilters "k8s.io/apiserver/pkg/server/filters"
)

// DeprecatedInsecureServingInfo is required to serve http.  HTTP does NOT include authentication or authorization.
// You shouldn't be using this.  It makes sig-auth sad.
// DeprecatedInsecureServingInfo *ServingInfo

// BuildInsecureHandlerChain sets up the server to listen to http. Should be removed.
func BuildInsecureHandlerChain(apiHandler http.Handler, c *server.Config, enableAuth bool) http.Handler {
	handler := apiHandler
	if enableAuth {
		handler = genericapifilters.WithAuthorization(apiHandler, c.Authorization.Authorizer, c.Serializer)
	}

	handler = genericapifilters.WithAudit(handler, c.AuditBackend, c.AuditPolicyRuleEvaluator, c.LongRunningFunc)

	if enableAuth {
		failedHandler := genericapifilters.Unauthorized(c.Serializer)
		failedHandler = genericapifilters.WithFailedAuthenticationAudit(failedHandler, c.AuditBackend, c.AuditPolicyRuleEvaluator)
		handler = genericapifilters.WithAuthentication(handler, c.Authentication.Authenticator,
			failedHandler, c.Authentication.APIAudiences, c.Authentication.RequestHeaderConfig)
	} else {
		handler = genericapifilters.WithAuthentication(handler, server.InsecureSuperuser{}, nil, nil, nil)
	}

	handler = genericfilters.WithCORS(handler, c.CorsAllowedOriginList, nil, nil, nil, "true")
	handler = genericfilters.WithTimeoutForNonLongRunningRequests(handler, c.LongRunningFunc)
	handler = genericfilters.WithMaxInFlightLimit(handler, c.MaxRequestsInFlight, c.MaxMutatingRequestsInFlight, c.LongRunningFunc)
	handler = genericfilters.WithWaitGroup(handler, c.LongRunningFunc, c.NonLongRunningRequestWaitGroup)
	handler = genericapifilters.WithRequestInfo(handler, server.NewRequestInfoResolver(c))
	handler = genericfilters.WithPanicRecovery(handler, c.RequestInfoResolver)

	return handler
}

// type InsecureServingInfo struct {
// 	// BindAddress is the ip:port to serve on
// 	BindAddress string
// 	// BindNetwork is the type of network to bind to - defaults to "tcp", accepts "tcp",
// 	// "tcp4", and "tcp6".
// 	BindNetwork string
// }

// func (s *InsecureServingInfo) NewLoopbackClientConfig() (*rest.Config, error) {
// 	if s == nil {
// 		return nil, nil
// 	}

// 	host, port, err := server.LoopbackHostPort(s.BindAddress)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &rest.Config{
// 		Host: "http://" + net.JoinHostPort(host, port),
// 		// Increase QPS limits. The client is currently passed to all admission plugins,
// 		// and those can be throttled in case of higher load on apiserver - see #22340 and #22422
// 		// for more details. Once #22422 is fixed, we may want to remove it.
// 		QPS:   50,
// 		Burst: 100,
// 	}, nil
// }

// // NonBlockingRun spawns the insecure http server. An error is
// // returned if the ports cannot be listened on.
// func NonBlockingRun(insecureServingInfo *InsecureServingInfo, insecureHandler http.Handler, shutDownTimeout time.Duration, stopCh <-chan struct{}) error {
// 	// Use an internal stop channel to allow cleanup of the listeners on error.
// 	internalStopCh := make(chan struct{})

// 	if insecureServingInfo != nil && insecureHandler != nil {
// 		if err := serveInsecurely(insecureServingInfo, insecureHandler, shutDownTimeout, internalStopCh); err != nil {
// 			close(internalStopCh)
// 			return err
// 		}
// 	}

// 	// Now that the listener has bound successfully, it is the
// 	// responsibility of the caller to close the provided channel to
// 	// ensure cleanup.
// 	go func() {
// 		<-stopCh
// 		close(internalStopCh)
// 	}()

// 	return nil
// }

// // serveInsecurely run the insecure http server. It fails only if the initial listen
// // call fails. The actual server loop (stoppable by closing stopCh) runs in a go
// // routine, i.e. serveInsecurely does not block.
// func serveInsecurely(insecureServingInfo *InsecureServingInfo, insecureHandler http.Handler, shutDownTimeout time.Duration, stopCh <-chan struct{}) error {
// 	insecureServer := &http.Server{
// 		Addr:           insecureServingInfo.BindAddress,
// 		Handler:        insecureHandler,
// 		MaxHeaderBytes: 1 << 20,
// 	}
// 	klog.Infof("Serving insecurely on %s", insecureServingInfo.BindAddress)
// 	ln, _, err := options.CreateListener(insecureServingInfo.BindNetwork, insecureServingInfo.BindAddress)
// 	if err != nil {
// 		return err
// 	}
// 	err = server.RunServer(insecureServer, ln, shutDownTimeout, stopCh)
// 	return err
// }

// insecureSuperuser implements authenticator.Request to always return a superuser.
// This is functionally equivalent to skipping authentication and authorization,
// but allows apiserver code to stop special-casing a nil user to skip authorization checks.
// type insecureSuperuser struct{}

// func (insecureSuperuser) AuthenticateRequest(req *http.Request) (user.Info, bool, error) {
// 	return &user.DefaultInfo{
// 		Name:   "system:unsecured",
// 		Groups: []string{user.SystemPrivilegedGroup, user.AllAuthenticated},
// 	}, true, nil
// }
