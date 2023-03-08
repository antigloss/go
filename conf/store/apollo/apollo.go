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

// Package apollo implements Store for reading and watching configurations from Apollo.
package apollo

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/magiconair/properties"
	apollo "github.com/taptap/go-apollo"

	"github.com/antigloss/go/conf/store"
)

// New creates a Store object for reading and watching configurations from Apollo.
// Unspecified Apollo client options could be read from ENV.
//
//	Relations of Apollo client options and ENV keys:
//	  - URL:         APOLLO_META . It not found in ENV, default is http://apollo.meta
//	  - AppID:       APOLLO_APP_ID
//	  - Cluster:     APOLLO_CLUSTER . It not found in ENV, default is `default`
//	  - AccessKey:   APOLLO_ACCESS_KEY , if not found, then APOLLO_ACCESSKEY_SECRET
//	  - Namespaces:  APOLLO_NAMESPACE . Comma separated. For example: ns1,ns2,ns3. It not found in ENV, default is `application`
func New(opts ...option) store.Store {
	a := &apolloStore{
		unwatchCh: make(chan int),
	}
	a.opts.apply(opts...)
	return a
}

type apolloStore struct {
	opts      options
	client    apollo.Apollo
	watchOnce sync.Once
	unwatchCh chan int
}

// Load reads configurations from Apollo
func (a *apolloStore) Load() ([]store.ConfigContent, error) {
	err := a.opts.validate()
	if err != nil {
		return nil, err
	}

	a.client, err = apollo.New(a.opts.addr, a.opts.appID, apollo.AutoFetchOnCacheMiss(), apollo.Cluster(a.opts.cluster),
		apollo.AccessKey(a.opts.accessKey), apollo.PreloadNamespaces(a.opts.namespaces...))
	if err != nil {
		return nil, err
	}

	contents := make([]store.ConfigContent, len(a.opts.namespaces), len(a.opts.namespaces)+1)
	for i, ns := range a.opts.namespaces {
		contents[i].Type, err = store.ConfigType(ns)
		if err != nil {
			return nil, err
		}

		contents[i].Content, err = a.nsToContent(ns, contents[i].Type)
		if err != nil {
			return nil, err
		}
	}

	if a.opts.local == nil {
		return contents, nil
	}

	localConf := a.opts.local.conf[a.opts.appID]
	if len(localConf) == 0 {
		return contents, nil
	}

	props := properties.NewProperties()
	for key, val := range localConf {
		_, _, err = props.Set(key, val)
		if err != nil {
			return nil, err
		}
	}

	buff := bytes.NewBuffer(nil)
	_, err = props.WriteComment(buff, "#", properties.UTF8)
	if err != nil {
		return nil, err
	}
	contents = append(contents, store.ConfigContent{Type: store.ConfigTypeDefault, Content: buff.Bytes()})

	return contents, nil
}

// Watch watches configuration changes from Apollo
func (a *apolloStore) Watch(ch chan<- *store.ConfigChanges) error {
	if !a.opts.watch {
		return nil
	}

	if a.client == nil {
		return fmt.Errorf("`Load()` must be called before `Watch()`")
	}

	a.watchOnce.Do(func() {
		_ = a.client.Start()
		watchCh := a.client.Watch()

		go func() {
			for {
				select {
				case resp := <-watchCh:
					confType, err := store.ConfigType(resp.Namespace)
					if err != nil {
						continue
					}

					changes := &store.ConfigChanges{
						Config: store.ConfigContent{Type: confType},
					}
					changes.Config.Content, _ = a.confToContent(resp.NewValue, resp.Namespace, confType)
					if changes.Config.Content == nil {
						continue
					}

					for _, change := range resp.Changes {
						c := store.ConfigChange{Key: change.Key}
						if change.Type == apollo.ChangeTypeUpdate {
							c.Type = store.ChangeTypeUpdated
						} else if change.Type == apollo.ChangeTypeDelete {
							c.Type = store.ChangeTypeDeleted
						}
						changes.Changes = append(changes.Changes, c)
					}

					ch <- changes
				case <-a.unwatchCh:
					return
				}
			}
		}()
	})

	return nil
}

// Unwatch stops watching
func (a *apolloStore) Unwatch() {
	a.client.Stop()
	close(a.unwatchCh)
}

func (a *apolloStore) nsToContent(ns, confType string) ([]byte, error) {
	return a.confToContent(a.client.GetNameSpace(ns), ns, confType)
}

func (a *apolloStore) confToContent(conf apollo.Configurations, ns, confType string) ([]byte, error) {
	if len(conf) == 0 {
		return nil, fmt.Errorf("empty apollo conf. addr=%s app=%s cluster=%s ns=%s", a.opts.addr, a.opts.appID, a.opts.cluster, ns)
	}

	var cont []byte
	var err error

	switch confType {
	case store.ConfigTypeDefault:
		cont, err = propsToContent(conf)
	case store.ConfigTypeJSON, store.ConfigTypeYAML, store.ConfigTypeYML:
		cont, err = getContent(conf)
	default:
		err = fmt.Errorf("unsupported configuration type")
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %s-%s", err.Error(), a.opts.appID, ns)
	}

	if a.opts.tData != nil {
		cont, err = a.opts.tData.Replace(cont)
		if err != nil {
			return nil, fmt.Errorf("%s: %s-%s", err.Error(), a.opts.appID, ns)
		}
	}

	return cont, nil
}

const (
	bulkConfigKey = "content"
)

func propsToContent(conf map[string]interface{}) ([]byte, error) {
	p := properties.NewProperties()
	for key, val := range conf {
		_, _, err := p.Set(key, val.(string))
		if err != nil {
			return nil, err
		}
	}

	buff := bytes.NewBuffer(nil)
	_, err := p.WriteComment(buff, "#", properties.UTF8)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func getContent(conf map[string]interface{}) ([]byte, error) {
	data, ok := conf[bulkConfigKey]
	if !ok || data == nil || len(conf) != 1 {
		return nil, fmt.Errorf("invalid configuration content")
	}

	cont, ok := data.(string)
	if !ok {
		return nil, fmt.Errorf("invalid configuration content")
	}

	return []byte(cont), nil
}
