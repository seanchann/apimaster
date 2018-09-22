package options

import (
	"fmt"

	"k8s.io/apiserver/pkg/storage/storagebackend"

	"github.com/spf13/pflag"
)

type MongoDBOptions struct {
	StorageConfig *storagebackend.MongoExtendConfig
}

// func NewMongoDBOptions() *MongoDBOptions {
// 	mongo := &MongoDBOptions{
// 		StorageConfig: &s.StorageConfig.Mongodb,
// 	}
// 	s.MongoDB = mongo

// 	return mongo
// }

func (s *MongoDBOptions) Validate() []error {
	allErrors := []error{}
	if len(s.StorageConfig.ServerList) == 0 {
		allErrors = append(allErrors, fmt.Errorf("--mongo-servers must be specified"))
	}
	return allErrors
}

// AddMongoDBStorageFlags adds flags related to mysql storage for a specific APIServer to the specified FlagSet
func (s *MongoDBOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringSliceVar(&s.StorageConfig.ServerList, "mongo-servers", s.StorageConfig.ServerList, ""+
		"specify server to connented backend.eg:mongodb://myuser:mypass@localhost:40001,otherhost:40001/mydb.")

	fs.StringSliceVar(&s.StorageConfig.AdminCred, "mongo-admin", s.StorageConfig.AdminCred, ""+
		"specify admin cred eg:admindb,admin,123456.")

	fs.StringSliceVar(&s.StorageConfig.GeneralCred, "mongo-user", s.StorageConfig.GeneralCred, ""+
		"specify general cred for project eg:test,testUser,123456.")
}
