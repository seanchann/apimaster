module github.com/seanchann/apimaster

require (
	github.com/Azure/go-autorest v10.6.2+incompatible // indirect
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/Sirupsen/logrus v1.2.0 // indirect
	github.com/aws/aws-sdk-go v1.15.73
	github.com/beorn7/perks v0.0.0-20180321164747-3a771d992973 // indirect
	github.com/coreos/go-systemd v0.0.0-20181031085051-9002847aa142 // indirect
	github.com/denisenkom/go-mssqldb v0.0.0-20181014144952-4e0d7dc8888f // indirect
	github.com/docker/docker v1.13.1 // indirect
	github.com/elazarl/go-bindata-assetfs v1.0.0 // indirect
	github.com/emicklei/go-restful v2.8.0+incompatible
	github.com/emicklei/go-restful-swagger12 v0.0.0-20170926063155-7524189396c6 // indirect
	github.com/erikstmartin/go-testdb v0.0.0-20160219214506-8d10e4a1bae5 // indirect
	github.com/evanphx/json-patch v4.1.0+incompatible // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-openapi/spec v0.0.0-20180213232550-1de3e0542de6
	github.com/gogo/protobuf v1.1.1 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/btree v0.0.0-20180813153112-4030bb1f1f0c // indirect
	github.com/google/go-cmp v0.2.0 // indirect
	github.com/google/gofuzz v0.0.0-20170612174753-24818f796faf // indirect
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/gophercloud/gophercloud v0.0.0-20181208025821-26de66c23d78 // indirect
	github.com/gregjones/httpcache v0.0.0-20180305231024-9cad4c3443a7 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/hashicorp/golang-lru v0.5.0 // indirect
	github.com/imdario/mergo v0.3.6 // indirect
	github.com/jinzhu/gorm v1.9.1
	github.com/jinzhu/now v0.0.0-20181116074157-8ec929ed50c3 // indirect
	github.com/json-iterator/go v1.1.5 // indirect
	github.com/lib/pq v1.0.0 // indirect
	github.com/mattn/go-sqlite3 v1.10.0 // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/natefinch/lumberjack v2.0.0+incompatible // indirect
	github.com/onsi/ginkgo v1.7.0 // indirect
	github.com/onsi/gomega v1.4.3 // indirect
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/prometheus/client_golang v0.9.1 // indirect
	github.com/prometheus/client_model v0.0.0-20180712105110-5c3871d89910 // indirect
	github.com/prometheus/common v0.0.0-20181109100915-0b1957f9d949 // indirect
	github.com/prometheus/procfs v0.0.0-20181005140218-185b4288413d // indirect
	github.com/sirupsen/logrus v1.2.0 // indirect
	github.com/spf13/pflag v1.0.3
	golang.org/x/crypto v0.0.0-20181203042331-505ab145d0a9 // indirect
	golang.org/x/net v0.0.0-20181207154023-610586996380
	golang.org/x/oauth2 v0.0.0-20181203162652-d668ce993890 // indirect
	golang.org/x/sync v0.0.0-20181108010431-42b317875d0f // indirect
	golang.org/x/sys v0.0.0-20181206074257-70b957f3b65e // indirect
	golang.org/x/time v0.0.0-20181108054448-85acf8d2951c // indirect
	google.golang.org/appengine v1.3.0 // indirect
	google.golang.org/genproto v0.0.0-20181202183823-bd91e49a0898 // indirect
	google.golang.org/grpc v1.17.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/mgo.v2 v2.0.0-20180705113604-9856a29383ce
	gopkg.in/yaml.v1 v1.0.0-20140924161607-9f9df34309c0 // indirect
	k8s.io/api v0.0.0-20181108192910-6560175f5b2d // indirect
	k8s.io/apimachinery v0.0.0-20181108192626-90473842928c
	k8s.io/apiserver v0.0.0-20181109033751-c8132c133cb6
	k8s.io/client-go v9.0.0+incompatible
	sigs.k8s.io/yaml v1.1.0 // indirect
)

replace (
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.2.0
	k8s.io/apiserver => github.com/seanchann/apiserver v1.0.2
)
