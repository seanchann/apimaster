/********************************************************************
* Copyright (c) All Rights Reserved.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*******************************************************************/
package login

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"testing"
)

func TestAuthenticationErrorStatus(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		code       int
		reason     metav1.StatusReason
	}{
		{"login", NewLoginAuthError().Status(), http.StatusUnauthorized, metav1.StatusReasonUnauthorized},
		{"logout", NewLogoutError().Status(), http.StatusForbidden, metav1.StatusReasonForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var status metav1.Status
			require.NoError(t, json.Unmarshal([]byte(tc.body), &status))
			require.EqualValues(t, tc.code, status.Code)
			require.Equal(t, tc.reason, status.Reason)
		})
	}
}
