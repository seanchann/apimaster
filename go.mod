module github.com/seanchann/apimaster

require (
	github.com/aws/aws-sdk-go v1.15.73
	github.com/coreos/go-systemd v0.0.0-20181031085051-9002847aa142 // indirect
	github.com/denisenkom/go-mssqldb v0.0.0-20190418034912-35416408c946 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/emicklei/go-restful v2.8.0+incompatible
	github.com/erikstmartin/go-testdb v0.0.0-20160219214506-8d10e4a1bae5 // indirect
	github.com/evanphx/json-patch v4.1.0+incompatible // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-openapi/spec v0.17.2
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/imdario/mergo v0.3.6 // indirect
	github.com/jinzhu/gorm v1.9.1
	github.com/jinzhu/now v1.0.0 // indirect
	github.com/json-iterator/go v1.1.5 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/lib/pq v1.1.0 // indirect
	github.com/mattn/go-sqlite3 v1.10.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20190414153302-2ae31c8b6b30 // indirect
	github.com/spf13/pflag v1.0.3
	golang.org/x/net v0.0.0-20190311183353-d8887717615a
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/mgo.v2 v2.0.0-20180705113604-9856a29383ce
	k8s.io/apimachinery v0.0.0-20190425132440-17f84483f500
	k8s.io/apiserver v0.0.0-20181109033751-c8132c133cb6
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/component-base v0.0.0-20190424053038-9fe063da3132
)

replace k8s.io/apiserver => github.com/seanchann/apiserver v1.0.3
