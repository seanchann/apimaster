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

package login

import (
	"encoding/json"
	"fmt"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type errLoginAuthFailed struct {
	message string
}

func NewLoginAuthError() errLoginAuthFailed {
	return errLoginAuthFailed{message: "user or password is incorrect"}
}

func (e errLoginAuthFailed) Error() string {
	return fmt.Sprintf("%v", e.message)
}

func (e errLoginAuthFailed) Status() string {
	status := metav1.Status{
		Status:  metav1.StatusFailure,
		Code:    http.StatusForbidden,
		Reason:  metav1.StatusReasonForbidden,
		Message: e.Error(),
	}

	statusstr, _ := json.Marshal(status)

	return string(statusstr)
}

type errLogoutFailed struct {
	message string
}

func NewLogoutError() errLoginAuthFailed {
	return errLoginAuthFailed{message: "user or token is incorrect"}
}

func (e errLogoutFailed) Error() string {
	return fmt.Sprintf("%v", e.message)
}

func (e errLogoutFailed) Status() string {
	status := metav1.Status{
		Status:  metav1.StatusFailure,
		Code:    http.StatusUnprocessableEntity,
		Reason:  metav1.StatusReasonInvalid,
		Message: e.Error(),
	}

	statusstr, _ := json.Marshal(status)

	return string(statusstr)

}
