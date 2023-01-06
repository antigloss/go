package file

import "github.com/antigloss/go/conf/tdata"

// ConfigPath 配置文件路径
type ConfigPath struct {
	Path      string // 配置文件路径，可以是具体的文件，也可以是目录。如果路径是目录，则会读取目录下的所有文件（不包括 . 开头的文件）
	Recursive bool   // true 表示递归读取所有子目录的文件，false 则只读取 Path 指定目录的文件
}

// WithConfigPaths 设置配置文件路径。如果路径是目录，则会读取目录下的所有文件（不包括 . 开头的文件）
func WithConfigPaths(paths ...ConfigPath) option {
	return func(o *options) {
		o.paths = paths
	}
}

// WithTemplateData 开启模板替换功能，使用 tData 替换配置中的模板参数
func WithTemplateData(tData tdata.TemplateData) option {
	return func(o *options) {
		o.tData = tData
	}
}

type option func(options *options)

type options struct {
	paths []ConfigPath
	tData tdata.TemplateData
}

func (o *options) apply(opts ...option) {
	for _, opt := range opts {
		opt(o)
	}
}
