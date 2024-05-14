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

package informers

import (
	reflect "reflect"
	time "time"

	clientset "aml.io/rtc/pkg/client/clientset"
	internalInform "aml.io/rtc/pkg/client/generated/informers"
	iam "aml.io/rtc/pkg/client/generated/informers/iam"
	internalinterfaces "aml.io/rtc/pkg/client/generated/informers/internalinterfaces"

	masterinform "github.com/seanchann/apimaster/pkg/client/informers"
	coreres "github.com/seanchann/apimaster/pkg/client/informers/coreres"
	masterinternalinterfaces "github.com/seanchann/apimaster/pkg/client/informers/internalinterfaces"
	rbac "github.com/seanchann/apimaster/pkg/client/informers/rbac"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	cache "k8s.io/client-go/tools/cache"
)

// TweakListOptionsFunc is a function that transforms a v1.ListOptions.
type TweakListOptionsFunc func(*v1.ListOptions)

// SharedInformerOption defines the functional option type for SharedInformerFactory.
type SharedInformerOption func(*sharedInformerFactory) (internalInform.SharedInformerOption, masterinform.SharedInformerOption)

// NewInformerFunc takes clientset.Interface and time.Duration to return a SharedIndexInformer.
type NewInformerFunc func(clientset.Interface, time.Duration) cache.SharedIndexInformer

type sharedInformerFactory struct {
	masterSharedInformerFactory   masterinform.SharedInformerFactory
	internalSharedInformerFactory internalInform.SharedInformerFactory

	//store common options for internal and master informers
	namespace        string
	transform        cache.TransformFunc
	tweakListOptions TweakListOptionsFunc
	customResync     map[reflect.Type]time.Duration
}

// WithCustomResyncConfig sets a custom resync period for the specified informer types.
func WithCustomResyncConfig(resyncConfig map[v1.Object]time.Duration) SharedInformerOption {
	return func(factory *sharedInformerFactory) (internalInform.SharedInformerOption, masterinform.SharedInformerOption) {
		for k, v := range resyncConfig {
			factory.customResync[reflect.TypeOf(k)] = v
		}
		return internalInform.WithCustomResyncConfig(resyncConfig), masterinform.WithCustomResyncConfig(resyncConfig)
	}
}

// WithTweakListOptions sets a custom filter on all listers of the configured SharedInformerFactory.
func WithTweakListOptions(tweakListOptions TweakListOptionsFunc) SharedInformerOption {
	return func(factory *sharedInformerFactory) (internalInform.SharedInformerOption, masterinform.SharedInformerOption) {
		factory.tweakListOptions = tweakListOptions
		internalTweakListOptions := func(options *v1.ListOptions) {
			tweakListOptions(options)
		}
		masterinformTweakListOptions := func(options *v1.ListOptions) {
			tweakListOptions(options)
		}
		return internalInform.WithTweakListOptions(internalTweakListOptions), masterinform.WithTweakListOptions(masterinformTweakListOptions)
	}
}

// WithNamespace limits the SharedInformerFactory to the specified namespace.
func WithNamespace(namespace string) SharedInformerOption {
	return func(factory *sharedInformerFactory) (internalInform.SharedInformerOption, masterinform.SharedInformerOption) {
		factory.namespace = namespace
		return internalInform.WithNamespace(namespace), masterinform.WithNamespace(namespace)
	}
}

// WithTransform sets a transform on all informers.
func WithTransform(transform cache.TransformFunc) SharedInformerOption {
	return func(factory *sharedInformerFactory) (internalInform.SharedInformerOption, masterinform.SharedInformerOption) {
		factory.transform = transform
		return internalInform.WithTransform(transform), masterinform.WithTransform(transform)
	}
}

// NewSharedInformerFactory constructs a new instance of sharedInformerFactory for all namespaces.
func NewSharedInformerFactory(client clientset.Interface, defaultResync time.Duration) SharedInformerFactory {
	return NewSharedInformerFactoryWithOptions(client, defaultResync)
}

// NewFilteredSharedInformerFactory constructs a new instance of sharedInformerFactory.
// Listers obtained via this SharedInformerFactory will be subject to the same filters
// as specified here.
// Deprecated: Please use NewSharedInformerFactoryWithOptions instead
func NewFilteredSharedInformerFactory(client clientset.Interface, defaultResync time.Duration, namespace string, tweakListOptions TweakListOptionsFunc) SharedInformerFactory {
	return NewSharedInformerFactoryWithOptions(client, defaultResync, WithNamespace(namespace), WithTweakListOptions(tweakListOptions))
}

// NewSharedInformerFactoryWithOptions constructs a new instance of a SharedInformerFactory with additional options.
func NewSharedInformerFactoryWithOptions(client clientset.Interface, defaultResync time.Duration, options ...SharedInformerOption) SharedInformerFactory {

	// Apply all options
	internalOption := []internalInform.SharedInformerOption{}
	masterOption := []masterinform.SharedInformerOption{}
	for _, opt := range options {
		internalOpt, masterOpt := opt(&sharedInformerFactory{})
		internalOption = append(internalOption, internalOpt)
		masterOption = append(masterOption, masterOpt)
	}

	factory := &sharedInformerFactory{
		masterSharedInformerFactory:   masterinform.NewSharedInformerFactoryWithOptions(client, defaultResync, masterOption...),
		internalSharedInformerFactory: internalInform.NewSharedInformerFactoryWithOptions(client, defaultResync, internalOption...),
	}

	return factory
}

func (f *sharedInformerFactory) Start(stopCh <-chan struct{}) {
	f.internalSharedInformerFactory.Start(stopCh)
	f.masterSharedInformerFactory.Start(stopCh)
}

func (f *sharedInformerFactory) Shutdown() {
	f.internalSharedInformerFactory.Shutdown()
	f.masterSharedInformerFactory.Shutdown()
}

func (f *sharedInformerFactory) WaitForCacheSync(stopCh <-chan struct{}) map[reflect.Type]bool {
	res := f.internalSharedInformerFactory.WaitForCacheSync(stopCh)
	for k, v := range f.masterSharedInformerFactory.WaitForCacheSync(stopCh) {
		res[k] = v
	}

	return res
}

// InformerFor returns the SharedIndexInformer for obj using an internal
// client.
// internalinterfaces.NewInformerFunc
func (f *sharedInformerFactory) InformerFor(obj runtime.Object, newFunc interface{}) cache.SharedIndexInformer {
	objGVK := obj.GetObjectKind().GroupVersionKind()

	switch objGVK.Group {
	case "coreres":
		fallthrough
	case "rbac":
		return f.masterSharedInformerFactory.InformerFor(obj, newFunc.(masterinternalinterfaces.NewInformerFunc))
	default:
		return f.internalSharedInformerFactory.InformerFor(obj, newFunc.(internalinterfaces.NewInformerFunc))
	}

}

// SharedInformerFactory provides shared informers for resources in all known
// API group versions.
//
// It is typically used like this:
//
//	ctx, cancel := context.Background()
//	defer cancel()
//	factory := NewSharedInformerFactory(client, resyncPeriod)
//	defer factory.WaitForStop()    // Returns immediately if nothing was started.
//	genericInformer := factory.ForResource(resource)
//	typedInformer := factory.SomeAPIGroup().V1().SomeType()
//	factory.Start(ctx.Done())          // Start processing these informers.
//	synced := factory.WaitForCacheSync(ctx.Done())
//	for v, ok := range synced {
//	    if !ok {
//	        fmt.Fprintf(os.Stderr, "caches failed to sync: %v", v)
//	        return
//	    }
//	}
//
//	// Creating informers can also be created after Start, but then
//	// Start must be called again:
//	anotherGenericInformer := factory.ForResource(resource)
//	factory.Start(ctx.Done())
type SharedInformerFactory interface {
	// Start initializes all requested informers. They are handled in goroutines
	// which run until the stop channel gets closed.
	Start(stopCh <-chan struct{})

	// Shutdown marks a factory as shutting down. At that point no new
	// informers can be started anymore and Start will return without
	// doing anything.
	//
	// In addition, Shutdown blocks until all goroutines have terminated. For that
	// to happen, the close channel(s) that they were started with must be closed,
	// either before Shutdown gets called or while it is waiting.
	//
	// Shutdown may be called multiple times, even concurrently. All such calls will
	// block until all goroutines have terminated.
	Shutdown()

	// WaitForCacheSync blocks until all started informers' caches were synced
	// or the stop channel gets closed.
	WaitForCacheSync(stopCh <-chan struct{}) map[reflect.Type]bool

	// ForResource gives generic access to a shared informer of the matching type.
	ForResource(resource schema.GroupVersionResource) (GenericInformer, error)

	// InformerFor returns the SharedIndexInformer for obj using an internal
	// client.
	InformerFor(obj runtime.Object, newFunc interface{}) cache.SharedIndexInformer

	Coreres() coreres.Interface
	Rbac() rbac.Interface

	Iam() iam.Interface
}

func (f *sharedInformerFactory) Iam() iam.Interface {
	return f.internalSharedInformerFactory.Iam()
}

func (f *sharedInformerFactory) Coreres() coreres.Interface {
	return f.masterSharedInformerFactory.Coreres()
}

func (f *sharedInformerFactory) Rbac() rbac.Interface {
	return f.masterSharedInformerFactory.Rbac()
}
