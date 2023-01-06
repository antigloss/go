/*
 *
 * Copyright (C) 2023 Antigloss Huang (https://github.com/antigloss) All rights reserved.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package apollo

import (
	"fmt"
	"os"

	"github.com/antigloss/go/conf/tdata"
)

// WithURL sets Apollo URL
func WithURL(url string) option {
	return func(o *options) {
		o.addr = url
	}
}

// WithAppID sets AppID to read configurations from
func WithAppID(id string) option {
	return func(o *options) {
		o.appID = id
	}
}

// WithCluster sets Cluster to read configurations from
func WithCluster(cluster string) option {
	return func(o *options) {
		o.cluster = cluster
	}
}

// WithAccessKey sets Apollo Access Key
func WithAccessKey(ak string) option {
	return func(o *options) {
		o.accessKey = ak
	}
}

// WithNamespaces sets namespaces to read configurations from
func WithNamespaces(namespaces ...string) option {
	return func(o *options) {
		o.namespaces = namespaces
	}
}

// WithTemplateData sets template data source.
// Will use configurations from `tData` to replace templates in the configurations read from Apollo
func WithTemplateData(tData tdata.TemplateData) option {
	return func(o *options) {
		o.tData = tData
	}
}

// WithLocalConfig sets LocalConfig to read Apollo Access Key from.
// Will also use configurations from LocalConfig to override configurations read from Apollo
func WithLocalConfig(cfg *localConfig) option {
	return func(o *options) {
		o.local = cfg
	}
}

// EnableWatch enables watching configuration changes
func EnableWatch() option {
	return func(o *options) {
		o.watch = true
	}
}

const (
	envAddr       = "APOLLO_META"
	envAppID      = "APOLLO_APP_ID"
	envCluster    = "APOLLO_CLUSTER"
	envAccessKey  = "APOLLO_ACCESS_KEY"
	envAccessKey2 = "APOLLO_ACCESSKEY_SECRET"
	envNamespace  = "APOLLO_NAMESPACE"

	defaultAddr      = "http://apollo.meta"
	defaultCluster   = "default"
	defaultNamespace = "application"
)

type option func(options *options)

type options struct {
	addr       string
	appID      string
	cluster    string
	accessKey  string
	namespaces []string
	local      *localConfig
	tData      tdata.TemplateData
	watch      bool
}

func (o *options) apply(opts ...option) {
	for _, opt := range opts {
		opt(o)
	}

	if o.addr == "" {
		if v := os.Getenv(envAddr); v != "" {
			o.addr = v
		} else {
			o.addr = defaultAddr
		}
	}

	if o.appID == "" {
		if v := os.Getenv(envAppID); v != "" {
			o.appID = v
		}
	}

	if o.cluster == "" {
		if v := os.Getenv(envCluster); v != "" {
			o.cluster = v
		} else {
			o.cluster = defaultCluster
		}
	}

	if o.accessKey == "" {
		if v := os.Getenv(envAccessKey); v != "" {
			o.accessKey = v
		} else if v = os.Getenv(envAccessKey2); v != "" {
			o.accessKey = v
		}
	}

	if len(o.namespaces) == 0 {
		if v := os.Getenv(envNamespace); v != "" {
			o.namespaces = []string{v}
		} else {
			o.namespaces = []string{defaultNamespace}
		}
	}

	if o.accessKey == "" && o.appID != "" && o.local != nil {
		if apolloKeys := o.local.conf["apollo-keys"]; apolloKeys != nil {
			o.accessKey = apolloKeys[o.appID]
		}
	}
}

func (o *options) validate() error {
	if o.appID == "" {
		return fmt.Errorf("AppID not specified")
	}
	return nil
}
