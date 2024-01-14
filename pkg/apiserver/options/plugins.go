/*
Copyright 2014 The Kubernetes Authors.

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

package options

// This file exists to force the desired plugin implementations to be linked.
// This should probably be part of some configuration fed into the build for a
// given binary target.
import (
	// Admission policies
	"github.com/seanchann/apimaster/plugin/pkg/admission/admit"
	"github.com/seanchann/apimaster/plugin/pkg/admission/deny"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/admission"
)

// AllOrderedPlugins is the list of all the plugins in order.
var AllOrderedPlugins = []string{
	admit.PluginName, // AlwaysAdmit
	deny.PluginName,  // AlwaysDeny
}

// RegisterAllAdmissionPlugins registers all admission plugins and
// sets the recommended plugins order.
func RegisterAllAdmissionPlugins(plugins *admission.Plugins) {
	admit.Register(plugins) // DEPRECATED as no real meaning
	deny.Register(plugins)  // DEPRECATED as no real meaning
}

// DefaultOffAdmissionPlugins get admission plugins off by default for apimaster.
func DefaultOffAdmissionPlugins() sets.String {
	defaultOnPlugins := sets.NewString()

	return sets.NewString(AllOrderedPlugins...).Difference(defaultOnPlugins)
}
