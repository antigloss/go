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

// Package store defines the Store interface, some common types and some common functions.
package store

// Store is the interface from which configurations can be read and watched
type Store interface {
	Load() ([]ConfigContent, error)       // read configurations
	Watch(ch chan<- *ConfigChanges) error // watch configuration changes
	Unwatch()                             // stop watching
}

// ConfigContent configuration content read from a Store object
type ConfigContent struct {
	Type    string // configuration format: json, yaml, properties...
	Content []byte // configuration content
}

// ChangeType is the change type of configuration
type ChangeType int

const (
	ChangeTypeAdded   = iota // 新增配置项
	ChangeTypeUpdated        // 修改配置项
	ChangeTypeDeleted        // 删除配置项
)

// ConfigChange change of configuration
type ConfigChange struct {
	Type ChangeType // 变化类型
	Key  string     // 变化的配置项
}

// ConfigChanges changes of configurations
type ConfigChanges struct {
	Config  ConfigContent  // 配置内容
	Changes []ConfigChange // 配置项的变化
}
