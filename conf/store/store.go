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

// ConfigContent is the configuration content read from a Store object
type ConfigContent struct {
	Type    string // configuration format: json, yaml, properties...
	Content []byte // configuration content
}

// ChangeType is the change type of configuration
type ChangeType int

const (
	ChangeTypeAdded   = iota // configuration added
	ChangeTypeUpdated        // configuration updated
	ChangeTypeDeleted        // configuration deleted
)

// ConfigChange change of configuration
type ConfigChange struct {
	Type ChangeType
	Key  string // key of the changed configuration
}

// ConfigChanges changes of configurations
type ConfigChanges struct {
	Config  ConfigContent
	Changes []ConfigChange
}
