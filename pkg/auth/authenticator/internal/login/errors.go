/********************************************************************
* Copyright (c) 2008 - 2024. sean <seandev@foxmail>
* All rights reserved.
*
* PROPRIETARY RIGHTS of the following material in either
* electronic or paper format pertain to sean.
* All manufacturing, reproduction, use, and sales involved with
* this subject MUST conform to the license agreement signed
* with sean.
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
