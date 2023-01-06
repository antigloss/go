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
