package apollo

import (
	"fmt"
	"os"

	"github.com/antigloss/go/conf/tdata"
)

// WithURL 设置 Apollo 地址
func WithURL(url string) option {
	return func(o *options) {
		o.addr = url
	}
}

// WithAppID 设置从哪个 AppID 读取配置
func WithAppID(id string) option {
	return func(o *options) {
		o.appID = id
	}
}

// WithCluster 设置从哪个 Cluster 读取配置
func WithCluster(cluster string) option {
	return func(o *options) {
		o.cluster = cluster
	}
}

// WithAccessKey 设置 AccessKey
func WithAccessKey(ak string) option {
	return func(o *options) {
		o.accessKey = ak
	}
}

// WithNamespaces 设置从哪些 namespaces 读取配置
func WithNamespaces(namespaces ...string) option {
	return func(o *options) {
		o.namespaces = namespaces
	}
}

// WithTemplateData 开启模板替换功能，使用 tData 替换配置中的模板参数
func WithTemplateData(tData tdata.TemplateData) option {
	return func(o *options) {
		o.tData = tData
	}
}

// WithLocalConfig 设置本地配置，从中读取 Apollo 秘钥，并使用本地配置覆盖 Apollo 上的配置
func WithLocalConfig(cfg *localConfig) option {
	return func(o *options) {
		o.local = cfg
	}
}

// EnableWatch 允许监听配置变化
func EnableWatch() option {
	return func(o *options) {
		o.watch = true
	}
}

const (
	// 默认从环境变量读取的值
	envAddr       = "APOLLO_META"
	envAppID      = "APOLLO_APP_ID"
	envCluster    = "APOLLO_CLUSTER"
	envAccessKey  = "APOLLO_ACCESS_KEY"
	envAccessKey2 = "APOLLO_ACCESSKEY_SECRET"
	envNamespace  = "APOLLO_NAMESPACE"

	// 默认值
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
