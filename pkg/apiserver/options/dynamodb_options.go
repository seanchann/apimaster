package options

import (
	"fmt"

	"k8s.io/apiserver/pkg/storage/storagebackend"

	"github.com/spf13/pflag"
)

type DynamoDBOptions struct {
	StorageConfig *storagebackend.AWSDynamoDBConfig
}

// func NewDynamoDBOptions() *DynamoDBOptions {
// 	dynamo := &DynamoDBOptions{
// 		StorageConfig: &s.StorageConfig.AWSDynamoDB,
// 	}
// 	s.Dynamodb = dynamo

// 	return dynamo
// }

func (s *DynamoDBOptions) Validate() []error {
	allErrors := []error{}
	if len(s.StorageConfig.Region) == 0 {
		allErrors = append(allErrors, fmt.Errorf("--aws-region must be specified"))
	}
	return allErrors
}

// AddMysqlStorageFlags adds flags related to mysql storage for a specific APIServer to the specified FlagSet
func (s *DynamoDBOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.StorageConfig.Region, "aws-region", s.StorageConfig.Region, ""+
		"specify the region where the session to connection.")

	fs.StringVar(&s.StorageConfig.Table, "aws-table", s.StorageConfig.Table, ""+
		"specify the table name. default(program name)")

	fs.StringVar(&s.StorageConfig.Token, "aws-cred-token", s.StorageConfig.Table, ""+
		"specify the token for credentials.")

	fs.StringVar(&s.StorageConfig.AccessID, "aws-cred-accessid", s.StorageConfig.AccessID, ""+
		"specify the access id for credentials.")

	fs.StringVar(&s.StorageConfig.AccessKey, "aws-cred-accesskey", s.StorageConfig.AccessKey, ""+
		"specify the access key for credentials.")
}
