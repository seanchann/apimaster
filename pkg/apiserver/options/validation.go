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

package options

import "fmt"

// Validate checks the selected backend and application options without opening a database.
func (o *APIMasterOptions) Validate() []error {
	if o == nil {
		return []error{fmt.Errorf("apimaster options are required")}
	}
	var errors []error
	switch o.Backend {
	case StorageBackendTypeSqlite:
		if o.Sqlite == nil {
			errors = append(errors, fmt.Errorf("SQLite options are required"))
		} else {
			errors = append(errors, o.Sqlite.Validate()...)
		}
	case StorageBackendTypeMysql:
		if o.Mysql == nil {
			errors = append(errors, fmt.Errorf("MySQL options are required"))
		} else {
			errors = append(errors, o.Mysql.Validate()...)
		}
	case StorageBackendTypeEtcd:
		if o.Etcd == nil {
			errors = append(errors, fmt.Errorf("etcd options are required"))
		} else {
			errors = append(errors, o.Etcd.Validate()...)
		}
	default:
		errors = append(errors, fmt.Errorf("unknown storage backend %q", o.Backend))
	}
	errors = append(errors, o.GenericServerRunOptions.Validate()...)
	errors = append(errors, o.SecureServing.Validate()...)
	errors = append(errors, o.InsecureServing.Validate()...)
	errors = append(errors, o.Authentication.Validate()...)
	errors = append(errors, o.Authorization.Validate()...)
	errors = append(errors, o.Audit.Validate()...)
	errors = append(errors, o.Features.Validate()...)
	errors = append(errors, o.APIEnablement.Validate()...)
	errors = append(errors, o.Admission.Validate()...)
	return errors
}
