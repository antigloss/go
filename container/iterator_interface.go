/*
 *
 * lomap - Linked Ordered Map, an ordered map that supports iteration in insertion order.
 * Copyright (C) 2016 Antigloss Huang (https://github.com/antigloss) All rights reserved.
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

// Package container provides some goroutine-unsafe containers and some goroutine-safe containers (under the container/concurrent folder).
package container

// Iterator is an interface for iterators of any container type.
type Iterator interface {
	// IsValid returns true if the iterator is valid for use, false otherwise.
	// We must not call Next, Key, or Value if IsValid returns false.
	IsValid() bool
	// Next advances the iterator to the next element of the map
	Next()
	// Value returns the value of the underlying element
	Value() interface{}
}

// MapIterator is a common interface for iterators of maps.
type MapIterator interface {
	Iterator
	// Key returns the key of the underlying element
	Key() interface{}
}
