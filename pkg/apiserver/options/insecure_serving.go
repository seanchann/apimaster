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

package options

import (
	"fmt"
	"net"
	"strconv"

	"github.com/spf13/pflag"
	"k8s.io/apiserver/pkg/server"
)

// DeprecatedInsecureServingOptions retains the donor localhost serving configuration.
type DeprecatedInsecureServingOptions struct {
	BindAddress net.IP
	BindPort    int
	BindNetwork string
}

// AddFlags exposes the donor insecure serving flags.
func (s *DeprecatedInsecureServingOptions) AddFlags(fs *pflag.FlagSet) {
	if s == nil {
		return
	}
	fs.IPVar(&s.BindAddress, "insecure-bind-address", s.BindAddress, "The localhost address for HTTP serving.")
	fs.IntVar(&s.BindPort, "insecure-port", s.BindPort, "The localhost HTTP port; zero disables HTTP serving.")
}

// Validate restricts insecure serving to localhost and a valid port.
func (s *DeprecatedInsecureServingOptions) Validate() []error {
	if s == nil || s.BindPort == 0 {
		return nil
	}
	if s.BindPort < 0 || s.BindPort > 65535 || !s.BindAddress.IsLoopback() {
		return []error{fmt.Errorf("insecure serving requires a loopback address and valid port")}
	}
	return nil
}

// ApplyTo binds a localhost listener, whose owner must close it on construction failure.
func (s *DeprecatedInsecureServingOptions) ApplyTo(info **server.DeprecatedInsecureServingInfo) error {
	if s == nil || s.BindPort == 0 {
		return nil
	}
	if errs := s.Validate(); len(errs) != 0 {
		return errs[0]
	}
	network := s.BindNetwork
	if network == "" {
		network = "tcp"
	}
	listener, err := net.Listen(network, net.JoinHostPort(s.BindAddress.String(), strconv.Itoa(s.BindPort)))
	if err != nil {
		return err
	}
	*info = &server.DeprecatedInsecureServingInfo{Listener: listener}
	return nil
}
