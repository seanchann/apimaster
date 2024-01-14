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

package initializer

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/apiserver/pkg/cel/openapi/resolver"
	quota "k8s.io/apiserver/pkg/quota/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/component-base/featuregate"
)

type WantsExternalNormalClientSet interface {
	SetExternalNormalClientSet(normalClientSet interface{})
	admission.InitializationValidator
}

// WantsExternalNormalInformerFactory defines a function which sets InformerFactory for admission plugins that need it
type WantsExternalNormalInformerFactory interface {
	SetExternalNormalInformerFactory(normalInformer interface{})
	admission.InitializationValidator
}

// WantsAuthorizer defines a function which sets Authorizer for admission plugins that need it.
type WantsAuthorizer interface {
	SetAuthorizer(authorizer.Authorizer)
	admission.InitializationValidator
}

// WantsQuotaConfiguration defines a function which sets quota configuration for admission plugins that need it.
type WantsQuotaConfiguration interface {
	SetQuotaConfiguration(quota.Configuration)
	admission.InitializationValidator
}

// WantsDrainedNotification defines a function which sets the notification of where the apiserver
// has already been drained for admission plugins that need it.
// After receiving that notification, Admit/Validate calls won't be called anymore.
type WantsDrainedNotification interface {
	SetDrainedNotification(<-chan struct{})
	admission.InitializationValidator
}

// WantsFeatureGate defines a function which passes the featureGates for inspection by an admission plugin.
// Admission plugins should not hold a reference to the featureGates.  Instead, they should query a particular one
// and assign it to a simple bool in the admission plugin struct.
//
//	func (a *admissionPlugin) InspectFeatureGates(features featuregate.FeatureGate){
//	    a.myFeatureIsOn = features.Enabled("my-feature")
//	}
type WantsFeatures interface {
	InspectFeatureGates(featuregate.FeatureGate)
	admission.InitializationValidator
}

type WantsDynamicClient interface {
	SetDynamicClient(dynamic.Interface)
	admission.InitializationValidator
}

// WantsRESTMapper defines a function which sets RESTMapper for admission plugins that need it.
type WantsRESTMapper interface {
	SetRESTMapper(meta.RESTMapper)
	admission.InitializationValidator
}

// WantsSchemaResolver defines a function which sets the SchemaResolver for
// an admission plugin that needs it.
type WantsSchemaResolver interface {
	SetSchemaResolver(resolver resolver.SchemaResolver)
	admission.InitializationValidator
}
