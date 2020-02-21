module github.com/seanchann/apimaster

go 1.13

require (
	contrib.go.opencensus.io/exporter/ocagent v0.4.12 // indirect
	git.apache.org/thrift.git v0.12.0 // indirect
	github.com/Azure/go-autorest v12.0.0+incompatible // indirect
	github.com/aws/aws-sdk-go v1.29.4
	github.com/emicklei/go-restful v2.9.5+incompatible
	github.com/go-openapi/spec v0.19.3
	github.com/golang/lint v0.0.0-20180702182130-06c8688daad7 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/jinzhu/gorm v1.9.12
	github.com/spf13/pflag v1.0.5
	golang.org/x/build v0.0.0-20190314133821-5284462c4bec // indirect
	golang.org/x/net v0.0.0-20200202094626-16171245cfb2
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22
	k8s.io/apiextensions-apiserver v0.17.3 // indirect
	k8s.io/apimachinery v0.17.3
	k8s.io/apiserver v0.17.3
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/component-base v0.17.3
	k8s.io/klog v1.0.0
	k8s.io/kubernetes v1.14.1
)

replace (
	golang.org/x/sys => golang.org/x/sys v0.0.0-20190813064441-fde4db37ae7a // pinned to release-branch.go1.13
	golang.org/x/tools => golang.org/x/tools v0.0.0-20190821162956-65e3620a7ae7 // pinned to release-branch.go1.13
	k8s.io/api => k8s.io/api v0.0.0-20200131193051-d9adff57e763
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20200131192631-731dcecc2054
	k8s.io/apiserver => github.com/seanchann/apiserver v1.17.0
	k8s.io/client-go => k8s.io/client-go v0.0.0-20200131194155-0cdd283dfd7a
	k8s.io/component-base => k8s.io/component-base v0.0.0-20200131194811-85b325a9731b
)
