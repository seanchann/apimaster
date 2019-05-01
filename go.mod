module github.com/seanchann/apimaster

require (
	contrib.go.opencensus.io/exporter/ocagent v0.4.12 // indirect
	github.com/Azure/go-autorest v12.0.0+incompatible // indirect
	github.com/aws/aws-sdk-go v1.19.14
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/emicklei/go-restful v2.9.3+incompatible
	github.com/erikstmartin/go-testdb v0.0.0-20160219214506-8d10e4a1bae5 // indirect
	github.com/go-openapi/spec v0.19.0
	k8s.io/klog v0.0.0-20160126235308-23def4e6c14b
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
	golang.org/x/sync => golang.org/x/sync v0.0.0-20181108010431-42b317875d0f
	golang.org/x/sys => golang.org/x/sys v0.0.0-20190209173611-3b5209105503
	golang.org/x/tools => golang.org/x/tools v0.0.0-20190313210603-aa82965741a9
	k8s.io/api => k8s.io/api v0.0.0-20190425012535-181e1f9c52c1
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190425132440-17f84483f500
	k8s.io/apiserver => github.com/seanchann/apiserver v1.0.4
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190425172711-65184652c889
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190419212335-ff26e7842f9d
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190424053038-9fe063da3132
)
