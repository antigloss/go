package lomap

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
var nbs sort.IntSlice

func TestLinkedRBTree(tt *testing.T) {
	t = tt
	rand.Seed(time.Now().Unix())

	rbt := New(func(a, b interface{}) int {
		return a.(int) - b.(int)
	})

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

func insertRandomly(rbt *LinkedOrderedMap, insertedNums sort.IntSlice, m map[int]int) {
	i := 0
	for i != kInsertTimes {
		n := rand.Int()
		rbt.Set(n, n)

		_, found := m[n]
		if found {
			continue
		}

		insertedNums[len(m)] = n
		m[n] = n
		i++
	}
}

func removeRandomly(rbt *LinkedOrderedMap, insertedNums, deletedNums sort.IntSlice, m map[int]int, deleteTimes int) {
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
			rbt.Remove(deletedNums[j])
		}
	}
}

func runTestCases(msg string, rbt *LinkedOrderedMap, m map[int]int, insertedNums sort.IntSlice) bool {
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

func verifySize(msg string, rbt *LinkedOrderedMap, m map[int]int, insertedNums sort.IntSlice) bool {
	if len(m) != rbt.Size() || len(insertedNums) != rbt.Size() {
		t.Errorf("%s. Unexpected number of elements! mLen=%d iLen=%d rbtSize=%d",
			msg, len(m), len(insertedNums), rbt.Size())
		return false
	}
	return true
}

func verifyData(msg string, rbt *LinkedOrderedMap, m map[int]int) bool {
	for k, v := range m {
		val, found := rbt.Get(k)
		if !found || val.(int) != v {
			t.Errorf("%s. Get() failed! Expecting %d but gets %v", msg, k, val)
			return false
		}
	}
	return true
}

func verifyInsertOrder(msg string, rbt *LinkedOrderedMap, insertedNums sort.IntSlice) bool {
	i := 0
	for it := rbt.LinkedIterator(); it.IsValid(); it.Next() {
		if insertedNums[i] != it.Value().(int) {
			t.Errorf("%s. Wrong insert order! Expecting %d but gets %d", insertedNums[i], it.Value().(int))
			return false
		}
		i++
	}

	i = len(insertedNums) - 1
	for it := rbt.ReverseLinkedIterator(); it.IsValid(); it.Next() {
		if insertedNums[i] != it.Value().(int) {
			t.Errorf("%s. Wrong insert order! Expecting %d but gets %d", insertedNums[i], it.Value().(int))
			return false
		}
		i--
	}

	return true
}

func verifySortedOrder(msg string, rbt *LinkedOrderedMap, insertedNums sort.IntSlice) bool {
	var sortedNums sort.IntSlice
	sortedNums = append(sortedNums, insertedNums...)
	sortedNums.Sort()

	i := 0
	for it := rbt.Iterator(); it.IsValid(); it.Next() {
		if sortedNums[i] != it.Value().(int) {
			t.Errorf("%s. Ordered iteration %d: Expecting %d but gets %d", msg, i, sortedNums[i], it.Value().(int))
			return false
		}
		i++
	}

	i = len(sortedNums) - 1
	for it := rbt.ReverseIterator(); it.IsValid(); it.Next() {
		if sortedNums[i] != it.Value().(int) {
			t.Errorf("%s. Reverse ordered iteration %d: Expecting %d but gets %d", msg, i, sortedNums[i], it.Value().(int))
			return false
		}
		i--
	}

	return true
}
