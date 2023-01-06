package env

import "github.com/antigloss/go/conf/tdata"

// WithTemplateData 开启模板替换功能，使用 tData 替换配置中的模板参数
func WithTemplateData(tData tdata.TemplateData) option {
	return func(o *options) {
		o.tData = tData
	}
}

type option func(options *options)

type options struct {
	tData tdata.TemplateData
}

func (o *options) apply(opts ...option) {
	for _, opt := range opts {
		opt(o)
	}
}
