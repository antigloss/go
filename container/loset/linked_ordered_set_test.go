/*
 *
 * loset - Linked Ordered Set, an ordered set that supports iteration in insertion order.
 * Copyright (C) 2022 Antigloss Huang (https://github.com/antigloss) All rights reserved.
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

package loset

import (
	"math/rand"
	"sort"
	"testing"
	"time"
)

const (
	kInsertTimes = 200000
)

var t *testing.T

func TestLinkedOrderedSet(tt *testing.T) {
	t = tt
	rand.Seed(time.Now().Unix())

	rbt := New[int]()

	// Phase 1

	// Insert random keys and values
	m := map[int]int{}
	insertedNums := make(sort.IntSlice, kInsertTimes, kInsertTimes*2)
	insertRandomly(rbt, insertedNums, m)

	if !runTestCases("After insertion 1", rbt, m, insertedNums) {
		return
	}

	// Prepare keys to be removed
	deleteTimes := len(insertedNums) / 2
	deletedNums := make(sort.IntSlice, deleteTimes, kInsertTimes*2)
	removeRandomly(rbt, insertedNums, deletedNums, m, deleteTimes)
	insertedNums = insertedNums[0 : len(insertedNums)-deleteTimes]

	if !runTestCases("After deletion 1", rbt, m, insertedNums) {
		return
	}

	// Phase 2

	insertedNums = insertedNums[0 : len(insertedNums)+kInsertTimes]
	insertRandomly(rbt, insertedNums, m)

	if !runTestCases("After insertion 2", rbt, m, insertedNums) {
		return
	}

	deleteTimes = len(insertedNums) - 1
	deletedNums = deletedNums[0:deleteTimes]
	removeRandomly(rbt, insertedNums, deletedNums, m, deleteTimes)
	insertedNums = insertedNums[0 : len(insertedNums)-deleteTimes]

	if !runTestCases("After deletion 2", rbt, m, insertedNums) {
		return
	}

	// Phase 3

	deleteTimes = len(insertedNums)
	deletedNums = deletedNums[0:deleteTimes]
	removeRandomly(rbt, insertedNums, deletedNums, m, deleteTimes)
	insertedNums = insertedNums[0 : len(insertedNums)-deleteTimes]

	if !runTestCases("After deletion 3", rbt, m, insertedNums) {
		return
	}

	// Phase 4

	insertedNums = insertedNums[0 : len(insertedNums)+kInsertTimes]
	insertRandomly(rbt, insertedNums, m)

	if !runTestCases("After insertion 4", rbt, m, insertedNums) {
		return
	}

	deleteTimes = len(insertedNums)
	deletedNums = deletedNums[0:deleteTimes]
	removeRandomly(rbt, insertedNums, deletedNums, m, deleteTimes)
	insertedNums = insertedNums[0 : len(insertedNums)-deleteTimes]

	if !runTestCases("After deletion 4", rbt, m, insertedNums) {
		return
	}
}

func insertRandomly(rbt *LinkedOrderedSet[int], insertedNums sort.IntSlice, m map[int]int) {
	i := 0
	for i != kInsertTimes {
		n := rand.Int()
		rbt.Insert(n)

		_, found := m[n]
		if found {
			continue
		}

		insertedNums[len(m)] = n
		m[n] = n
		i++
	}
}

func removeRandomly(rbt *LinkedOrderedSet[int], insertedNums, deletedNums sort.IntSlice, m map[int]int, deleteTimes int) {
	for i := 0; i != deleteTimes; i++ {
		nLen := len(insertedNums)
		idx := rand.Int() % nLen
		deletedNums[i] = insertedNums[idx]
		delete(m, deletedNums[i])
		if idx+1 < nLen {
			copy(insertedNums[idx:], insertedNums[idx+1:])
		}
		insertedNums = insertedNums[0 : nLen-1]
	}

	// Remove every key twice to make sure no key will be removed mistakenly
	for i := 0; i != 2; i++ {
		for j := 0; j != deleteTimes; j++ {
			rbt.Erase(deletedNums[j])
		}
	}
}

func runTestCases(msg string, rbt *LinkedOrderedSet[int], m map[int]int, insertedNums sort.IntSlice) bool {
	if !verifySize(msg, rbt, m, insertedNums) {
		return false
	}

	if !verifyData(msg, rbt, m) {
		return false
	}

	if !verifyInsertOrder(msg, rbt, insertedNums) {
		return false
	}

	if !verifySortedOrder(msg, rbt, insertedNums) {
		return false
	}

	return true
}

func verifySize(msg string, rbt *LinkedOrderedSet[int], m map[int]int, insertedNums sort.IntSlice) bool {
	if len(m) != rbt.Size() || len(insertedNums) != rbt.Size() {
		t.Errorf("%s. Unexpected number of elements! mLen=%d iLen=%d rbtSize=%d",
			msg, len(m), len(insertedNums), rbt.Size())
		return false
	}
	return true
}

func verifyData(msg string, rbt *LinkedOrderedSet[int], m map[int]int) bool {
	for k := range m {
		if rbt.Count(k) != 1 {
			t.Errorf("%s. Count() failed! %d not found!", msg, k)
			return false
		}
	}
	return true
}

func verifyInsertOrder(msg string, rbt *LinkedOrderedSet[int], insertedNums sort.IntSlice) bool {
	i := 0
	for it := rbt.LinkedIterator(); it.IsValid(); it.Next() {
		if insertedNums[i] != it.Value() {
			t.Errorf("%s. Wrong insert order! Expecting %d but gets %d", msg, insertedNums[i], it.Value())
			return false
		}
		i++
	}

	i = len(insertedNums) - 1
	for it := rbt.ReverseLinkedIterator(); it.IsValid(); it.Next() {
		if insertedNums[i] != it.Value() {
			t.Errorf("%s. Wrong insert order! Expecting %d but gets %d", msg, insertedNums[i], it.Value())
			return false
		}
		i--
	}

	return true
}

func verifySortedOrder(msg string, rbt *LinkedOrderedSet[int], insertedNums sort.IntSlice) bool {
	var sortedNums sort.IntSlice
	sortedNums = append(sortedNums, insertedNums...)
	sortedNums.Sort()

	i := 0
	for it := rbt.Iterator(); it.IsValid(); it.Next() {
		if sortedNums[i] != it.Value() {
			t.Errorf("%s. Ordered iteration %d: Expecting %d but gets %d", msg, i, sortedNums[i], it.Value())
			return false
		}
		i++
	}

	i = len(sortedNums) - 1
	for it := rbt.ReverseIterator(); it.IsValid(); it.Next() {
		if sortedNums[i] != it.Value() {
			t.Errorf("%s. Reverse ordered iteration %d: Expecting %d but gets %d", msg, i, sortedNums[i], it.Value())
			return false
		}
		i--
	}

	return true
}
