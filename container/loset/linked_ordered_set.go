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

// Package loset implements a linked ordered set which supports iteration in insertion order.
// It's also optimized for ordered traverse. loset is short for Linked Ordered Set.
//
// Caution: This package is not goroutine-safe!
package loset

import "golang.org/x/exp/constraints"

// LinkedOrderedSet is a linked ordered set which supports iteration in insertion order.
// It's also optimized for ordered traverse.
type LinkedOrderedSet[K constraints.Ordered] struct {
	root        *lrbtNode[K] // root of the rbtree
	head        *lrbtNode[K] // head and tail forms an double linked list in insertion order
	tail        *lrbtNode[K]
	orderedHead *lrbtNode[K] // orderedHead and orderedTail forms an double linked list in ascend order
	orderedTail *lrbtNode[K]
	size        int // size of the set
}

// New is the only way to get a new, ready-to-use LinkedOrderedSet object.
//
// Example:
//
//	lom := New[int]()
func New[K constraints.Ordered]() *LinkedOrderedSet[K] {
	return &LinkedOrderedSet[K]{}
}

// Insert inserts a new element into the LinkedOrderedSet if it doesn't already exist.
// Nothing will be changed if the LinkedOrderedSet already contains an element with the specified value.
//
//	value: value to be inserted
//
// Return value: true if the insertion takes place and false otherwise.
func (m *LinkedOrderedSet[K]) Insert(value K) bool {
	return m.set(value)
}

// Erase removes the element with the given value from the set.
func (m *LinkedOrderedSet[K]) Erase(value K) {
	m.erase(m.search(value))
}

// Empty returns true if the set does not contain any element, otherwise it returns false.
func (m *LinkedOrderedSet[K]) Empty() bool {
	return m.size == 0
}

// Size returns the number of elements in the set.
func (m *LinkedOrderedSet[K]) Size() int {
	return m.size
}

// Iterator returns an iterator for iterating the LinkedOrderedSet.
func (m *LinkedOrderedSet[K]) Iterator() *Iterator[K] {
	return &Iterator[K]{m.orderedHead}
}

// ReverseIterator returns an iterator for iterating the LinkedOrderedSet in reverse order.
func (m *LinkedOrderedSet[K]) ReverseIterator() *ReverseIterator[K] {
	return &ReverseIterator[K]{m.orderedTail}
}

// LinkedIterator returns an iterator for iterating the LinkedOrderedSet in insertion order.
func (m *LinkedOrderedSet[K]) LinkedIterator() *LinkedIterator[K] {
	return &LinkedIterator[K]{m.head}
}

// FindLinkedIterator returns a LinkedIterator to the given `value`.
// If found, LinkedIterator.IsValid() returns true, otherwise it returns false.
func (m *LinkedOrderedSet[K]) FindLinkedIterator(value K) *LinkedIterator[K] {
	return &LinkedIterator[K]{m.search(value)}
}

// MoveToBack move the element specified by `iter` to the back of the linked list as if it is just inserted.
func (m *LinkedOrderedSet[K]) MoveToBack(iter *LinkedIterator[K]) {
	node := iter.node
	if node == nil || node.next == nil { // node is nil or the last node
		return
	}

	if node.prev != nil {
		node.prev.next = node.next
	} else {
		m.head = node.next
	}
	node.next.prev = node.prev
	node.prev = m.tail
	node.next = nil
	m.tail.next = node
	m.tail = node
}

// EraseByLinkedIterator erases the element specified by `iter`
func (m *LinkedOrderedSet[K]) EraseByLinkedIterator(iter *LinkedIterator[K]) {
	m.erase(iter.node)
	iter.node = nil
}

// EraseFront erases the front element
func (m *LinkedOrderedSet[K]) EraseFront() {
	m.erase(m.head)
}

// ReverseLinkedIterator returns an iterator for iterating the LinkedOrderedSet in reverse insertion order.
func (m *LinkedOrderedSet[K]) ReverseLinkedIterator() *ReverseLinkedIterator[K] {
	return &ReverseLinkedIterator[K]{m.tail}
}

// Clear removes all elements from the set.
func (m *LinkedOrderedSet[K]) Clear() {
	m.root = nil
	m.head = nil
	m.tail = nil
	m.orderedHead = nil
	m.orderedTail = nil
	m.size = 0
}

// Count returns the number of elements with given `value`, which is either 1 or 0 since this container does not allow duplicates.
//
//	value: value of the elements to count
func (m *LinkedOrderedSet[K]) Count(value K) int {
	if m.search(value) != nil {
		return 1
	}
	return 0
}

// set inserts a new node into the LinkedOrderedSet or updates the existing node with the new value.
func (m *LinkedOrderedSet[K]) set(key K) bool {
	newNode := &lrbtNode[K]{k: key}
	if m.root != nil {
		node := m.root
		for {
			if key > node.k { // k is bigger than the node.k, go right.
				if node.right != nil {
					node = node.right
				} else {
					node.right = newNode
					newNode.nodeType = kLRBTNodeTypeRightChild
					break
				}
			} else if key < node.k { // k is smaller than the node.k, go left.
				if node.left != nil {
					node = node.left
				} else {
					node.left = newNode
					newNode.nodeType = kLRBTNodeTypeLeftChild
					break
				}
			} else { // k already existed
				return false
			}
		}
		newNode.parent = node
		m.insertCase2(newNode)
		// insert ordered linked list
		newNode.prev = m.tail
		m.tail.next = newNode
		m.tail = newNode
		// ordered linked list
		if newNode.isLeftChild() {
			var nextNode *lrbtNode[K]
			if newNode.right == nil {
				nextNode = newNode.parent
			} else {
				nextNode = newNode.right.leftmostChild()
			}
			newNode.orderedPrev = nextNode.orderedPrev
			newNode.orderedNext = nextNode
			nextNode.orderedPrev = newNode
			if newNode.orderedPrev != nil {
				newNode.orderedPrev.orderedNext = newNode
			} else {
				m.orderedHead = newNode
			}
		} else if newNode.isRightChild() {
			var prevNode *lrbtNode[K]
			if newNode.left == nil {
				prevNode = newNode.parent
			} else {
				prevNode = newNode.left.rightmostChild()
			}
			newNode.orderedPrev = prevNode
			newNode.orderedNext = prevNode.orderedNext
			prevNode.orderedNext = newNode
			if newNode.orderedNext != nil {
				newNode.orderedNext.orderedPrev = newNode
			} else {
				m.orderedTail = newNode
			}
		} else {
			newNode.orderedPrev = newNode.left
			newNode.orderedNext = newNode.right
			newNode.left.orderedNext = newNode
			newNode.right.orderedPrev = newNode
		}
	} else {
		m.root = newNode
		m.head = newNode
		m.tail = newNode
		m.orderedHead = newNode
		m.orderedTail = newNode
		newNode.isBlack = true
		newNode.nodeType = kLRBTNodeTypeRoot
	}

	m.size++
	return true
}

// Case 1: root node
func (m *LinkedOrderedSet[K]) insertCase1(node *lrbtNode[K]) {
	if node.parent != nil {
		m.insertCase2(node)
	} else { // Root node
		node.isBlack = true
	}
}

// Case 2: black node can have children of any color
func (m *LinkedOrderedSet[K]) insertCase2(node *lrbtNode[K]) {
	if !node.parent.isBlack {
		m.insertCase3(node)
	}
}

// Case 3: red nodes' children must be black
func (m *LinkedOrderedSet[K]) insertCase3(node *lrbtNode[K]) {
	uncle := node.parent.sibling()
	if !uncle.isBlackNode() {
		node.parent.isBlack = true
		uncle.isBlack = true
		node.parent.parent.isBlack = false
		m.insertCase1(node.parent.parent)
	} else {
		m.insertCase4(node)
	}
}

// Case 4
func (m *LinkedOrderedSet[K]) insertCase4(node *lrbtNode[K]) {
	if node.isRightChild() && node.parent.isLeftChild() {
		m.rotateLeft(node.parent)
		node = node.left
	} else if node.isLeftChild() && node.parent.isRightChild() {
		m.rotateRight(node.parent)
		node = node.right
	}
	m.insertCase5(node)
}

// Case 5
func (m *LinkedOrderedSet[K]) insertCase5(node *lrbtNode[K]) {
	node.parent.isBlack = true
	grandparent := node.parent.parent
	grandparent.isBlack = false
	if node.isLeftChild() && node.parent.isLeftChild() {
		m.rotateRight(grandparent)
	} else if node.isRightChild() && node.parent.isRightChild() {
		m.rotateLeft(grandparent)
	}
}

// Case 1: root node
func (m *LinkedOrderedSet[K]) deleteCase1(node *lrbtNode[K]) {
	if node.parent != nil {
		m.deleteCase2(node)
	}
}

// Case 2: sibling node is red
func (m *LinkedOrderedSet[K]) deleteCase2(node *lrbtNode[K]) {
	sibling := node.sibling()
	if !sibling.isBlackNode() {
		node.parent.isBlack = false
		sibling.isBlack = true
		if node.isLeftChild() {
			m.rotateLeft(node.parent)
		} else {
			m.rotateRight(node.parent)
		}
	}
	m.deleteCase3(node)
}

// Case 3: parent, sibling and its children are black
func (m *LinkedOrderedSet[K]) deleteCase3(node *lrbtNode[K]) {
	sibling := node.sibling()
	if node.parent.isBlack && sibling.isBlack && sibling.left.isBlackNode() && sibling.right.isBlackNode() {
		sibling.isBlack = false
		m.deleteCase1(node.parent)
	} else {
		m.deleteCase4(node, sibling)
	}
}

// Case 4: parent is red and sibling and its children are black
func (m *LinkedOrderedSet[K]) deleteCase4(node, sibling *lrbtNode[K]) {
	if !node.parent.isBlack && sibling.isBlack && sibling.left.isBlackNode() && sibling.right.isBlackNode() {
		sibling.isBlack = false
		node.parent.isBlack = true
	} else {
		m.deleteCase5(node, sibling)
	}
}

// Case 5: only one child of sibling is red
func (m *LinkedOrderedSet[K]) deleteCase5(node, sibling *lrbtNode[K]) {
	if node.isLeftChild() && sibling.isBlack && !sibling.left.isBlackNode() && sibling.right.isBlackNode() {
		sibling.isBlack = false
		sibling.left.isBlack = true
		m.rotateRight(sibling)
	} else if node.isRightChild() && sibling.isBlack && !sibling.right.isBlackNode() && sibling.left.isBlackNode() {
		sibling.isBlack = false
		sibling.right.isBlack = true
		m.rotateLeft(sibling)
	}
	m.deleteCase6(node)
}

// Case 6
func (m *LinkedOrderedSet[K]) deleteCase6(node *lrbtNode[K]) {
	sibling := node.sibling()
	sibling.isBlack = node.parent.isBlack
	node.parent.isBlack = true
	if node.isLeftChild() && !sibling.right.isBlackNode() {
		sibling.right.isBlack = true
		m.rotateLeft(node.parent)
	} else if !sibling.left.isBlackNode() {
		sibling.left.isBlack = true
		m.rotateRight(node.parent)
	}
}

func (m *LinkedOrderedSet[K]) rotateLeft(node *lrbtNode[K]) {
	right := node.right
	m.replaceNode(node, right)
	node.right = right.left
	if right.left != nil {
		right.left.parent = node
		right.left.nodeType = kLRBTNodeTypeRightChild
	}
	right.left = node
	node.parent = right
	node.nodeType = kLRBTNodeTypeLeftChild
}

func (m *LinkedOrderedSet[K]) rotateRight(node *lrbtNode[K]) {
	left := node.left
	m.replaceNode(node, left)
	node.left = left.right
	if left.right != nil {
		left.right.parent = node
		left.right.nodeType = kLRBTNodeTypeLeftChild
	}
	left.right = node
	node.parent = left
	node.nodeType = kLRBTNodeTypeRightChild
}

func (m *LinkedOrderedSet[K]) search(key K) (node *lrbtNode[K]) {
	node = m.root
	for node != nil {
		if key > node.k {
			node = node.right
		} else if key < node.k {
			node = node.left
		} else {
			break
		}
	}
	return
}

func (m *LinkedOrderedSet[K]) replaceNode(oldNode *lrbtNode[K], newNode *lrbtNode[K]) {
	if oldNode.parent == nil {
		m.root = newNode
		if newNode != nil {
			newNode.nodeType = kLRBTNodeTypeRoot
		}
	} else {
		if oldNode.isLeftChild() {
			oldNode.parent.left = newNode
			if newNode != nil {
				newNode.nodeType = kLRBTNodeTypeLeftChild
			}
		} else {
			oldNode.parent.right = newNode
			if newNode != nil {
				newNode.nodeType = kLRBTNodeTypeRightChild
			}
		}
	}
	if newNode != nil {
		newNode.parent = oldNode.parent
	}
}

func (m *LinkedOrderedSet[K]) erase(node *lrbtNode[K]) {
	if node == nil {
		return
	}

	needFixList := true
	// If both of the left and right child exist
	if node.left != nil && node.right != nil {
		predecessor := node.left.rightmostChild()
		node.k = predecessor.k

		// Fix insert ordered linked list
		if node.prev != nil && node.prev != predecessor && node.next != predecessor {
			node.prev.next = node.next
			if node.next == nil {
				m.tail = node.prev
			}
		}
		if node.next != nil && node.next != predecessor && node.prev != predecessor {
			node.next.prev = node.prev
			if node.prev == nil {
				m.head = node.next
			}
		}
		if predecessor.prev != node {
			node.prev = predecessor.prev
			if predecessor.prev != nil {
				predecessor.prev.next = node
			} else {
				m.head = node
			}
		}
		if predecessor.next != node {
			node.next = predecessor.next
			if predecessor.next != nil {
				predecessor.next.prev = node
			} else {
				m.tail = node
			}
		}

		// Fix ordered linked list
		if node.orderedPrev != nil && node.orderedPrev != predecessor && node.orderedNext != predecessor {
			node.orderedPrev.orderedNext = node.orderedNext
			if node.orderedNext == nil {
				m.orderedTail = node.orderedPrev
			}
		}
		if node.orderedNext != nil && node.orderedNext != predecessor && node.orderedPrev != predecessor {
			node.orderedNext.orderedPrev = node.orderedPrev
			if node.orderedPrev == nil {
				m.orderedHead = node.orderedNext
			}
		}
		if predecessor.orderedPrev != node {
			node.orderedPrev = predecessor.orderedPrev
			if predecessor.orderedPrev != nil {
				predecessor.orderedPrev.orderedNext = node
			} else {
				m.orderedHead = node
			}
		}
		if predecessor.orderedNext != node {
			node.orderedNext = predecessor.orderedNext
			if predecessor.orderedNext != nil {
				predecessor.orderedNext.orderedPrev = node
			} else {
				m.orderedTail = node
			}
		}

		//  Now the node to be deleted becomes the predecessor
		node = predecessor
		needFixList = false
	}

	// At this point, it's certain that node has at most one children
	var child *lrbtNode[K]
	if node.right == nil {
		child = node.left
	} else {
		child = node.right
	}
	if node.isBlack {
		node.isBlack = child.isBlackNode()
		m.deleteCase1(node)
	}
	m.replaceNode(node, child)
	// If the node that was deleted is a root node
	if node.parent == nil && child != nil {
		child.isBlack = true
	}

	if needFixList {
		// Fix insert ordered linked list
		if node.prev != nil {
			node.prev.next = node.next
		} else {
			m.head = node.next
		}
		if node.next != nil {
			node.next.prev = node.prev
		} else {
			m.tail = node.prev
		}

		// Fix ordered linked list
		if node.orderedPrev != nil {
			node.orderedPrev.orderedNext = node.orderedNext
		} else {
			m.orderedHead = node.orderedNext
		}
		if node.orderedNext != nil {
			node.orderedNext.orderedPrev = node.orderedPrev
		} else {
			m.orderedTail = node.orderedPrev
		}
	}

	m.size--
}

// Iterator is used for iterating the LinkedOrderedSet.
type Iterator[K constraints.Ordered] struct {
	node *lrbtNode[K]
}

// IsValid returns true if the iterator is valid for use, false otherwise.
// We must not call Next, Key, or Value if IsValid returns false.
func (it *Iterator[K]) IsValid() bool {
	return it.node != nil
}

// Next advances the iterator to the next element of the set
func (it *Iterator[K]) Next() {
	it.node = it.node.orderedNext
}

// Value returns the value of the underlying element
func (it *Iterator[K]) Value() K {
	return it.node.k
}

// ReverseIterator is used for iterating the LinkedOrderedSet in reverse order.
type ReverseIterator[K constraints.Ordered] struct {
	node *lrbtNode[K]
}

// IsValid returns true if the iterator is valid for use, false otherwise.
// We must not call Next, Key, or Value if IsValid returns false.
func (it *ReverseIterator[K]) IsValid() bool {
	return it.node != nil
}

// Next advances the iterator to the next element of the set in reverse order
func (it *ReverseIterator[K]) Next() {
	it.node = it.node.orderedPrev
}

// Value returns the value of the underlying element
func (it *ReverseIterator[K]) Value() K {
	return it.node.k
}

// LinkedIterator is used for iterating the LinkedOrderedSet in insertion order.
type LinkedIterator[K constraints.Ordered] struct {
	node *lrbtNode[K]
}

// IsValid returns true if the iterator is valid for use, false otherwise.
// We must not call Next, Key, or Value if IsValid returns false.
func (it *LinkedIterator[K]) IsValid() bool {
	return it.node != nil
}

// Next advances the iterator to the next element of the set in insertion order
func (it *LinkedIterator[K]) Next() {
	it.node = it.node.next
}

// Value returns the value of the underlying element
func (it *LinkedIterator[K]) Value() K {
	return it.node.k
}

// ReverseLinkedIterator is used for iterating the LinkedOrderedSet in reverse insertion order.
type ReverseLinkedIterator[K constraints.Ordered] struct {
	node *lrbtNode[K]
}

// IsValid returns true if the iterator is valid for use, false otherwise.
// We must not call Next, Key, or Value if IsValid returns false.
func (it *ReverseLinkedIterator[K]) IsValid() bool {
	return it.node != nil
}

// Next advances the iterator to the next element of the set in reverse insertion order
func (it *ReverseLinkedIterator[K]) Next() {
	it.node = it.node.prev
}

// Value returns the value of the underlying element
func (it *ReverseLinkedIterator[K]) Value() K {
	return it.node.k
}

type lrbtNodeType byte

const (
	kLRBTNodeTypeRoot lrbtNodeType = iota
	kLRBTNodeTypeLeftChild
	kLRBTNodeTypeRightChild
)

type lrbtNode[K constraints.Ordered] struct {
	k           K
	isBlack     bool
	nodeType    lrbtNodeType
	left        *lrbtNode[K]
	right       *lrbtNode[K]
	parent      *lrbtNode[K]
	prev        *lrbtNode[K]
	next        *lrbtNode[K]
	orderedPrev *lrbtNode[K]
	orderedNext *lrbtNode[K]
}

func (node *lrbtNode[K]) sibling() *lrbtNode[K] {
	if node.parent != nil {
		if node.isLeftChild() {
			return node.parent.right
		}
		return node.parent.left
	}
	return nil
}

func (node *lrbtNode[K]) rightmostChild() *lrbtNode[K] {
	for node.right != nil {
		node = node.right
	}
	return node
}

func (node *lrbtNode[K]) leftmostChild() *lrbtNode[K] {
	for node.left != nil {
		node = node.left
	}
	return node
}

func (node *lrbtNode[K]) isBlackNode() bool {
	if node != nil {
		return node.isBlack
	}
	return true
}

func (node *lrbtNode[K]) isLeftChild() bool {
	return node.nodeType == kLRBTNodeTypeLeftChild
}

func (node *lrbtNode[K]) isRightChild() bool {
	return node.nodeType == kLRBTNodeTypeRightChild
}
