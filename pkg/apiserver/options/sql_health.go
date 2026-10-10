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
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/healthz"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	"k8s.io/apiserver/pkg/storage/storagebackend/factory"
)

type sqlRESTOptions struct {
	generic.RESTOptionsGetter
	health   factory.Prober
	ready    factory.Prober
	once     sync.Once
	closeErr error
}

// Close releases both owned probes; server destruction and failure paths may call it repeatedly.
func (s *sqlRESTOptions) Close() error {
	s.once.Do(func() { s.closeErr = errors.Join(s.health.Close(), s.ready.Close()) })
	return s.closeErr
}

func applySQLRESTOptions(c *server.Config, config storagebackend.Config, getter generic.RESTOptionsGetter) error {
	health, err := factory.CreateProber(config)
	if err != nil {
		return err
	}
	ready, err := factory.CreateProber(config)
	if err != nil {
		health.Close()
		return err
	}
	options := &sqlRESTOptions{RESTOptionsGetter: getter, health: health, ready: ready}
	healthTimeout := config.HealthcheckTimeout
	if healthTimeout == 0 {
		healthTimeout = storagebackend.DefaultHealthcheckTimeout
	}
	readyTimeout := config.ReadycheckTimeout
	if readyTimeout == 0 {
		readyTimeout = storagebackend.DefaultReadinessTimeout
	}
	if healthTimeout < 0 || readyTimeout < 0 {
		options.Close()
		return errors.New("SQL check deadlines must be positive")
	}
	check := func(probe factory.Prober, timeout time.Duration) func(*http.Request) error {
		return func(request *http.Request) error {
			ctx, cancel := context.WithTimeout(request.Context(), timeout)
			defer cancel()
			return probe.Probe(ctx)
		}
	}
	c.AddHealthChecks(healthz.NamedGroupedCheck("sql-storage", "sql-storage", check(health, healthTimeout)))
	c.AddReadyzChecks(healthz.NamedGroupedCheck("sql-storage-readiness", "sql-storage", check(ready, readyTimeout)))
	c.RESTOptionsGetter = options
	return nil
}
