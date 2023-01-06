package apollo

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/magiconair/properties"
	apollo "github.com/taptap/go-apollo"

	"github.com/antigloss/go/conf/store"
)

// New 创建从 Apollo 读取配置的 Store 对象。没有设置的 Apollo 参数，会默认通过环境变量获取。
//
//	参数和环境变量的对应关系如下：
//	  - URL:         APOLLO_META 。如果连环境变量中都没有，则默认为 http://apollo.meta
//	  - AppID:       APOLLO_APP_ID
//	  - Cluster:     APOLLO_CLUSTER 。如果连环境变量中都没有，则默认为 default
//	  - AccessKey:   先 APOLLO_ACCESS_KEY ，后 APOLLO_ACCESSKEY_SECRET ，取第一个非空值
//	  - Namespaces:  APOLLO_NAMESPACE 。可使用逗号分割，配置多个 namespace ，如：ns1,ns2,ns3 。如果连环境变量中都没有，则默认为 application
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

// Load 加载配置
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

// Watch 监听配置变化
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

// Unwatch 取消监听
func (a *apolloStore) Unwatch() {
	a.client.Stop()
	close(a.unwatchCh)
}

func (a *apolloStore) nsToContent(ns, confType string) ([]byte, error) {
	return a.confToContent(a.client.GetNameSpace(ns), ns, confType)
}

func (a *apolloStore) confToContent(conf apollo.Configurations, ns, confType string) ([]byte, error) {
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
	bulkConfigKey = "content" // 配置为 json、yaml、xml 等格式时，key 固定为 content
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
