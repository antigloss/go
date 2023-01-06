package tdata

import "github.com/antigloss/go/conf/store"

// WithStores 设置 TemplateData 的配置数据来源
func WithStores(stores ...store.Store) option {
	return func(o *options) {
		o.stores = stores
	}
}

type option func(opts *options)

type options struct {
	stores []store.Store
}

func (o *options) apply(opts ...option) {
	for _, opt := range opts {
		opt(o)
	}
}
