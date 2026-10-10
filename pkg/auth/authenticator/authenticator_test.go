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
package authenticator

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestInvalidTokenReturnsError(t *testing.T) {
	a := NewLoginAuth([]byte("test-only-signing-key"), time.Minute, nil)
	require.NotPanics(t, func() { info, err := a.Validate("malformed-token"); require.Error(t, err); require.Nil(t, info) })
}
