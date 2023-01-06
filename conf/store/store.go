package store

// Store 配置存储
type Store interface {
	Load() ([]ConfigContent, error)       // 加载配置
	Watch(ch chan<- *ConfigChanges) error // 监听配置变化
	Unwatch()                             // 取消监听
}

// ConfigContent 从 Store 中读取到的配置内容
type ConfigContent struct {
	Type    string // 配置类型：json、yaml、properties 等等
	Content []byte // 配置内容
}

// ChangeType 配置项变化类型
type ChangeType int

const (
	ChangeTypeAdded   = iota // 新增配置项
	ChangeTypeUpdated        // 修改配置项
	ChangeTypeDeleted        // 删除配置项
)

// ConfigChange 配置项的变化
type ConfigChange struct {
	Type ChangeType // 变化类型
	Key  string     // 变化的配置项
}

// ConfigChanges 配置的变化
type ConfigChanges struct {
	Config  ConfigContent  // 配置内容
	Changes []ConfigChange // 配置项的变化
}
