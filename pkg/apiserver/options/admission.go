/*
Copyright 2018 The Kubernetes Authors.

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

import (
	"fmt"
	"strings"

	"github.com/seanchann/apimaster/pkg/admission/initializer"
	"github.com/spf13/pflag"
	"k8s.io/client-go/dynamic"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/admission"
	admissionmetrics "k8s.io/apiserver/pkg/admission/metrics"
	apiserverapi "k8s.io/apiserver/pkg/apis/apiserver"
	apiserverapiv1 "k8s.io/apiserver/pkg/apis/apiserver/v1"
	apiserverapiv1alpha1 "k8s.io/apiserver/pkg/apis/apiserver/v1alpha1"
	"k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
	"k8s.io/component-base/featuregate"
)

var configScheme = runtime.NewScheme()

func init() {
	utilruntime.Must(apiserverapi.AddToScheme(configScheme))
	utilruntime.Must(apiserverapiv1alpha1.AddToScheme(configScheme))
	utilruntime.Must(apiserverapiv1.AddToScheme(configScheme))
}

// AdmissionOptions holds the admission options.
// It is a wrap of generic AdmissionOptions.
type AdmissionOptions struct {
	// GenericAdmission holds the generic admission options.
	GenericAdmission *genericoptions.AdmissionOptions
	// DEPRECATED flag, should use EnabledAdmissionPlugins and DisabledAdmissionPlugins.
	// They are mutually exclusive, specify both will lead to an error.
	PluginNames []string
}

// NewAdmissionOptions creates a new instance of AdmissionOptions
// Note:
//
//	In addition it calls RegisterAllAdmissionPlugins to register
//	all kube-apiserver admission plugins.
//
//	Provides the list of RecommendedPluginOrder that holds sane values
//	that can be used by servers that don't care about admission chain.
//	Servers that do care can overwrite/append that field after creation.
func NewAdmissionOptions() *AdmissionOptions {
	options := genericoptions.NewAdmissionOptions()
	// register all admission plugins
	RegisterAllAdmissionPlugins(options.Plugins)
	// set RecommendedPluginOrder
	options.RecommendedPluginOrder = AllOrderedPlugins
	// set DefaultOffPlugins
	options.DefaultOffPlugins = DefaultOffAdmissionPlugins()

	return &AdmissionOptions{
		GenericAdmission: options,
	}
}

// AddFlags adds flags related to admission for kube-apiserver to the specified FlagSet
func (a *AdmissionOptions) AddFlags(fs *pflag.FlagSet) {
	if a == nil {
		return
	}
	fs.StringSliceVar(&a.PluginNames, "admission-control", a.PluginNames, ""+
		"Admission is divided into two phases. "+
		"In the first phase, only mutating admission plugins run. "+
		"In the second phase, only validating admission plugins run. "+
		"The names in the below list may represent a validating plugin, a mutating plugin, or both. "+
		"The order of plugins in which they are passed to this flag does not matter. "+
		"Comma-delimited list of: "+strings.Join(a.GenericAdmission.Plugins.Registered(), ", ")+".")
	fs.MarkDeprecated("admission-control", "Use --enable-admission-plugins or --disable-admission-plugins instead. Will be removed in a future version.")
	fs.Lookup("admission-control").Hidden = false

	a.GenericAdmission.AddFlags(fs)
}

// Validate verifies flags passed to kube-apiserver AdmissionOptions.
// Kube-apiserver verifies PluginNames and then call generic AdmissionOptions.Validate.
func (a *AdmissionOptions) Validate() []error {
	if a == nil {
		return nil
	}
	var errs []error
	if a.PluginNames != nil &&
		(a.GenericAdmission.EnablePlugins != nil || a.GenericAdmission.DisablePlugins != nil) {
		errs = append(errs, fmt.Errorf("admission-control and enable-admission-plugins/disable-admission-plugins flags are mutually exclusive"))
	}

	registeredPlugins := sets.NewString(a.GenericAdmission.Plugins.Registered()...)
	for _, name := range a.PluginNames {
		if !registeredPlugins.Has(name) {
			errs = append(errs, fmt.Errorf("admission-control plugin %q is unknown", name))
		}
	}

	errs = append(errs, a.GenericAdmission.Validate()...)

	return errs
}

// ApplyTo adds the admission chain to the server configuration.
// Kube-apiserver just call generic AdmissionOptions.ApplyTo.
func (a *AdmissionOptions) ApplyTo(
	c *server.Config,
	informers interface{},
	normalClient interface{},
	dynamicClient dynamic.Interface,
	features featuregate.FeatureGate,
	pluginInitializers ...admission.PluginInitializer,
) error {
	if a == nil {
		return nil
	}

	if a.PluginNames != nil {
		// pass PluginNames to generic AdmissionOptions
		a.GenericAdmission.EnablePlugins, a.GenericAdmission.DisablePlugins = computePluginNames(a.PluginNames, a.GenericAdmission.RecommendedPluginOrder)
	}

	return a.apiserverApplyTo(c, informers, normalClient, dynamicClient, features, pluginInitializers...)
}

// explicitly disable all plugins that are not in the enabled list
func computePluginNames(explicitlyEnabled []string, all []string) (enabled []string, disabled []string) {
	return explicitlyEnabled, sets.NewString(all...).Difference(sets.NewString(explicitlyEnabled...)).List()
}

// this function instread of apiserver/server/options/admisson.go/ApplyTo
func (a *AdmissionOptions) apiserverApplyTo(
	c *server.Config,
	informers interface{},
	normalClient interface{},
	dynamicClient dynamic.Interface,
	features featuregate.FeatureGate,
	pluginInitializers ...admission.PluginInitializer,
) error {
	if a == nil || a.GenericAdmission == nil {
		return nil
	}

	// Admission depends on CoreAPI to set SharedInformerFactory and ClientConfig.
	if informers == nil {
		return fmt.Errorf("admission depends on a Kubernetes core API shared informer, it cannot be nil")
	}

	pluginNames := a.enabledPluginNames()

	pluginsConfigProvider, err := admission.ReadAdmissionConfiguration(pluginNames,
		a.GenericAdmission.ConfigFile, configScheme)
	if err != nil {
		return fmt.Errorf("failed to read plugin config: %v", err)
	}

	genericInitializer := initializer.New(normalClient,
		dynamicClient, informers, c.Authorization.Authorizer, features, c.DrainedNotify())
	initializersChain := admission.PluginInitializers{genericInitializer}
	initializersChain = append(initializersChain, pluginInitializers...)

	admissionChain, err := a.GenericAdmission.Plugins.NewFromPlugins(pluginNames,
		pluginsConfigProvider, initializersChain, a.GenericAdmission.Decorators)
	if err != nil {
		return err
	}

	c.AdmissionControl = admissionmetrics.WithStepMetrics(admissionChain)
	return nil
}

// enabledPluginNames makes use of RecommendedPluginOrder, DefaultOffPlugins,
// EnablePlugins, DisablePlugins fields
// to prepare a list of ordered plugin names that are enabled.
func (a *AdmissionOptions) enabledPluginNames() []string {
	allOffPlugins := append(a.GenericAdmission.DefaultOffPlugins.List(),
		a.GenericAdmission.DisablePlugins...)
	disabledPlugins := sets.NewString(allOffPlugins...)
	enabledPlugins := sets.NewString(a.GenericAdmission.EnablePlugins...)
	disabledPlugins = disabledPlugins.Difference(enabledPlugins)

	orderedPlugins := []string{}
	for _, plugin := range a.GenericAdmission.RecommendedPluginOrder {
		if !disabledPlugins.Has(plugin) {
			orderedPlugins = append(orderedPlugins, plugin)
		}
	}

	return orderedPlugins
}
