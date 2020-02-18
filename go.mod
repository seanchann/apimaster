module github.com/seanchann/apimaster

go 1.13

require (
	contrib.go.opencensus.io/exporter/ocagent v0.4.12 // indirect
	github.com/Azure/go-autorest v12.0.0+incompatible // indirect
	github.com/aws/aws-sdk-go v1.19.14
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/emicklei/go-restful v2.9.3+incompatible
	github.com/erikstmartin/go-testdb v0.0.0-20160219214506-8d10e4a1bae5 // indirect
	github.com/go-openapi/spec v0.19.0
	k8s.io/klog v1.0.0
	github.com/gophercloud/gophercloud v0.0.0-20190418141522-bb98932a7b3a // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/jinzhu/gorm v1.9.4
	github.com/jinzhu/now v1.0.0 // indirect
	github.com/spf13/pflag v1.0.3
	golang.org/x/net v0.0.0-20190419010253-1f3472d942ba
	gopkg.in/mgo.v2 v2.0.0-20180705113604-9856a29383ce
	k8s.io/apimachinery v0.0.0-20190425132440-17f84483f500
	k8s.io/apiserver v0.0.0-00010101000000-000000000000
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/component-base v0.0.0-20190424053038-9fe063da3132
	k8s.io/kubernetes v1.14.1 // indirect
)

replace (
	k8s.io/apiserver => github.com/seanchann/apiserver v1.1.0
	golang.org/x/sys => golang.org/x/sys v0.0.0-20190813064441-fde4db37ae7a // pinned to release-branch.go1.13
	golang.org/x/tools => golang.org/x/tools v0.0.0-20190821162956-65e3620a7ae7 // pinned to release-branch.go1.13
	k8s.io/api => k8s.io/api v0.0.0-20200131193051-d9adff57e763
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20200131192631-731dcecc2054
	k8s.io/client-go => k8s.io/client-go v0.0.0-20200131194155-0cdd283dfd7a
	k8s.io/component-base => k8s.io/component-base v0.0.0-20200131194811-85b325a9731b
)
