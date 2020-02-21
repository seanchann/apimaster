module github.com/seanchann/apimaster

go 1.13

require (
	github.com/emicklei/go-restful v2.9.5+incompatible
	github.com/go-openapi/spec v0.19.3
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.17.3 // indirect
	k8s.io/apimachinery v0.17.3
	k8s.io/apiserver v0.17.3
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/component-base v0.17.3
	k8s.io/klog v1.0.0
)

replace (
	github.com/Azure/go-autorest/autorest => github.com/Azure/go-autorest/autorest v0.9.0
	github.com/Azure/go-autorest/autorest/adal => github.com/Azure/go-autorest/autorest/adal v0.5.0
	github.com/Azure/go-autorest/autorest/date => github.com/Azure/go-autorest/autorest/date v0.1.0
	github.com/Azure/go-autorest/autorest/mocks => github.com/Azure/go-autorest/autorest/mocks v0.2.0
	github.com/Azure/go-autorest/autorest/to => github.com/Azure/go-autorest/autorest/to v0.2.0
	github.com/Azure/go-autorest/autorest/validation => github.com/Azure/go-autorest/autorest/validation v0.1.0
	github.com/Azure/go-autorest/logger => github.com/Azure/go-autorest/logger v0.1.0
	github.com/Azure/go-autorest/tracing => github.com/Azure/go-autorest/tracing v0.5.0
	golang.org/x/sys => golang.org/x/sys v0.0.0-20190813064441-fde4db37ae7a // pinned to release-branch.go1.13
	golang.org/x/tools => golang.org/x/tools v0.0.0-20190821162956-65e3620a7ae7 // pinned to release-branch.go1.13
	k8s.io/api => k8s.io/api v0.0.0-20200131193051-d9adff57e763
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20200131192631-731dcecc2054
	k8s.io/apiserver => github.com/seanchann/apiserver v1.17.0
	k8s.io/client-go => k8s.io/client-go v0.0.0-20200131194155-0cdd283dfd7a
	k8s.io/component-base => k8s.io/component-base v0.0.0-20200131194811-85b325a9731b
)
