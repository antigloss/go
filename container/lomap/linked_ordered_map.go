// Author: https://github.com/antigloss

// Package lomap implements an linked ordered map which supports iteration in insertion order.
// It's also optimized for ordered traverse. lomap is short for Linked Ordered Map.
//
// Caution: This package is not goroutine-safe!
package lomap

// Comparator compares a and b and returns:
//     0 if they are equal
//     < 0 if a < b
//     > 0 if a > b
type Comparator func(a, b interface{}) int

// LinkedOrderedMap is an linked ordered map which supports iteration in insertion order.
// It's also optimized for ordered traverse.
type LinkedOrderedMap struct {
	root        *lrbtNode // root of the rbtree
	head        *lrbtNode // head and tail forms an double linked list in insertion order
	tail        *lrbtNode
	orderedHead *lrbtNode // orderedHead and orderedTail forms an double linked list in ascend order
	orderedTail *lrbtNode
	size        int        // size of the map
	comp        Comparator // for comparing keys of the map
}

// New is the only way to get a new, ready-to-use LinkedOrderedMap object.
//
//   comparator: for comparing keys of the LinkedOrderedMap
//
// Example:
//
//	 lom := New(func(a, b interface{}) int {return a.(int) - b.(int)})
func New(comparator Comparator) *LinkedOrderedMap {
	return &LinkedOrderedMap{comp: comparator}
}

// Insert inserts a new element into the LinkedOrderedMap if it doesn't already contain an element with an equivalent key.
// Nothing will be changed if the LinkedOrderedMap already contains an element with an equivalent key.
// Key should adhere to the comparator's type assertion, otherwise it will panic.
//   key: key of the value to be inserted
//   value: value to be inserted
// Return value: true if the insertion took place and false otherwise.
func (m *LinkedOrderedMap) Insert(key, value interface{}) bool {
	return m.set(key, value, false)
}

// Set inserts a new element into the LinkedOrderedMap or updates the existing element with the new value.
// Key should adhere to the comparator's type assertion, otherwise it will panic.
//   key: key of the value to be inserted/updated
//   value: value to be inserted/updated
// Return value: true if the insertion took place and false if the update took place.
func (m *LinkedOrderedMap) Set(key, value interface{}) bool {
	return m.set(key, value, true)
}

// Get returns value of the key and true if the given key is found.
// If the given key is not found, it returns nil, false
// Key should adhere to the comparator's type assertion, otherwise it will panic.
func (m *LinkedOrderedMap) Get(key interface{}) ( /*value*/ interface{} /*found*/, bool) {
	node := m.search(key)
	if node != nil {
		return node.v, true
	}
	return nil, false
}

// Erase removes the element with the given key from the map.
// Key should adhere to the comparator's type assertion, otherwise it will panic.
func (m *LinkedOrderedMap) Erase(key interface{}) {
	m.erase(m.search(key))
}

// Empty returns true if the map does not contain any element, otherwise it returns false.
func (m *LinkedOrderedMap) Empty() bool {
	return m.size == 0
}

// Size returns the number of elements in the map.
func (m *LinkedOrderedMap) Size() int {
	return m.size
}

// Iterator returns an iterator for iterating the LinkedOrderedMap.
func (m *LinkedOrderedMap) Iterator() *Iterator {
	return &Iterator{m.orderedHead}
}

// ReverseIterator returns an iterator for iterating the LinkedOrderedMap in reverse order.
func (m *LinkedOrderedMap) ReverseIterator() *ReverseIterator {
	return &ReverseIterator{m.orderedTail}
}

// LinkedIterator returns an iterator for iterating the LinkedOrderedMap in insertion order.
func (m *LinkedOrderedMap) LinkedIterator() *LinkedIterator {
	return &LinkedIterator{m.head}
}

// FindLinkedIterator returns a LinkedIterator to the `key`, or nil if not found.
// Key should adhere to the comparator's type assertion, otherwise it will panic.
func (m *LinkedOrderedMap) FindLinkedIterator(key interface{}) *LinkedIterator {
	node := m.search(key)
	if node != nil {
		return &LinkedIterator{node}
	}
	return nil
}

// MoveToBack move the element specified by `iter` to the back of the linked list as if it is just inserted.
func (m *LinkedOrderedMap) MoveToBack(iter *LinkedIterator) {
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
func (m *LinkedOrderedMap) EraseByLinkedIterator(iter *LinkedIterator) {
	m.erase(iter.node)
	iter.node = nil
}

// EraseFront erases the front element
func (m *LinkedOrderedMap) EraseFront() {
	m.erase(m.head)
}

// ReverseLinkedIterator returns an iterator for iterating the LinkedOrderedMap in reverse insertion order.
func (m *LinkedOrderedMap) ReverseLinkedIterator() *ReverseLinkedIterator {
	return &ReverseLinkedIterator{m.tail}
}

// TODO how to support Clone()?

// Clear removes all elements from the map.
func (m *LinkedOrderedMap) Clear() {
	m.root = nil
	m.head = nil
	m.tail = nil
	m.orderedHead = nil
	m.orderedTail = nil
	m.size = 0
}

// Count returns the number of elements with key key, which is either 1 or 0 since this container does not allow duplicates.
//
//   key: key value of the elements to count
func (m *LinkedOrderedMap) Count(key interface{}) int {
	if m.search(key) != nil {
		return 1
	}
	return 0
}

// set inserts a new node into the LinkedOrderedMap or updates the existing node with the new value.
func (m *LinkedOrderedMap) set(key, value interface{}, updateIfExist bool) bool {
	newNode := &lrbtNode{k: key, v: value}
	if m.root != nil {
		node := m.root
		for {
			ret := m.comp(key, node.k)
			if ret > 0 { // k is bigger than the node.k, go right.
				if node.right != nil {
					node = node.right
				} else {
					node.right = newNode
					newNode.nodeType = kLRBTNodeTypeRightChild
					break
				}
			} else if ret < 0 { // k is smaller than the node.k, go left.
				if node.left != nil {
					node = node.left
				} else {
					node.left = newNode
					newNode.nodeType = kLRBTNodeTypeLeftChild
					break
				}
			} else { // k already exists, updates the value.
				if updateIfExist {
					node.k = key
					node.v = value
				}
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
			var nextNode *lrbtNode
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
			var prevNode *lrbtNode
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
func (m *LinkedOrderedMap) insertCase1(node *lrbtNode) {
	if node.parent != nil {
		m.insertCase2(node)
	} else { // Root node
		node.isBlack = true
	}
}

// Case 2: black node can have children of any color
func (m *LinkedOrderedMap) insertCase2(node *lrbtNode) {
	if !node.parent.isBlack {
		m.insertCase3(node)
	}
}

// Case 3: red nodes' children must be black
func (m *LinkedOrderedMap) insertCase3(node *lrbtNode) {
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
func (m *LinkedOrderedMap) insertCase4(node *lrbtNode) {
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
func (m *LinkedOrderedMap) insertCase5(node *lrbtNode) {
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
func (m *LinkedOrderedMap) deleteCase1(node *lrbtNode) {
	if node.parent != nil {
		m.deleteCase2(node)
	}
}

// Case 2: sibling node is red
func (m *LinkedOrderedMap) deleteCase2(node *lrbtNode) {
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
func (m *LinkedOrderedMap) deleteCase3(node *lrbtNode) {
	sibling := node.sibling()
	if node.parent.isBlack && sibling.isBlack && sibling.left.isBlackNode() && sibling.right.isBlackNode() {
		sibling.isBlack = false
		m.deleteCase1(node.parent)
	} else {
		m.deleteCase4(node, sibling)
	}
}

// Case 4: parent is red and sibling and its children are black
func (m *LinkedOrderedMap) deleteCase4(node, sibling *lrbtNode) {
	if !node.parent.isBlack && sibling.isBlack && sibling.left.isBlackNode() && sibling.right.isBlackNode() {
		sibling.isBlack = false
		node.parent.isBlack = true
	} else {
		m.deleteCase5(node, sibling)
	}
}

// Case 5: only one child of sibling is red
func (m *LinkedOrderedMap) deleteCase5(node, sibling *lrbtNode) {
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
func (m *LinkedOrderedMap) deleteCase6(node *lrbtNode) {
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

func (m *LinkedOrderedMap) rotateLeft(node *lrbtNode) {
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

func (m *LinkedOrderedMap) rotateRight(node *lrbtNode) {
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

func (m *LinkedOrderedMap) search(key interface{}) (node *lrbtNode) {
	node = m.root
	for node != nil {
		ret := m.comp(key, node.k)
		if ret > 0 {
			node = node.right
		} else if ret < 0 {
			node = node.left
		} else {
			break
		}
	}
	return
}

func (m *LinkedOrderedMap) replaceNode(oldNode *lrbtNode, newNode *lrbtNode) {
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

func (m *LinkedOrderedMap) erase(node *lrbtNode) {
	if node == nil {
		return
	}

	needFixList := true
	// If both of the left and right child exist
	if node.left != nil && node.right != nil {
		predecessor := node.left.rightmostChild()
		node.k = predecessor.k
		node.v = predecessor.v

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
	var child *lrbtNode
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

// Iterator is used for iterating the LinkedOrderedMap.
type Iterator struct {
	node *lrbtNode
}

// IsValid returns true if the iterator is valid for use, false otherwise.
// We must not call Next, Key, or Value if IsValid returns false.
func (it *Iterator) IsValid() bool {
	return it.node != nil
}

// Next advances the iterator to the next element of the map
func (it *Iterator) Next() {
	it.node = it.node.orderedNext
}

// Key returns the key of the underlying element
func (it *Iterator) Key() interface{} {
	return it.node.k
}

// Value returns the value of the underlying element
func (it *Iterator) Value() interface{} {
	return it.node.v
}

// ReverseIterator is used for iterating the LinkedOrderedMap in reverse order.
type ReverseIterator struct {
	node *lrbtNode
}

// IsValid returns true if the iterator is valid for use, false otherwise.
// We must not call Next, Key, or Value if IsValid returns false.
func (it *ReverseIterator) IsValid() bool {
	return it.node != nil
}

// Next advances the iterator to the next element of the map in reverse order
func (it *ReverseIterator) Next() {
	it.node = it.node.orderedPrev
}

// Key returns the key of the underlying element
func (it *ReverseIterator) Key() interface{} {
	return it.node.k
}

// Value returns the value of the underlying element
func (it *ReverseIterator) Value() interface{} {
	return it.node.v
}

// LinkedIterator is used for iterating the LinkedOrderedMap in insertion order.
type LinkedIterator struct {
	node *lrbtNode
}

// IsValid returns true if the iterator is valid for use, false otherwise.
// We must not call Next, Key, or Value if IsValid returns false.
func (it *LinkedIterator) IsValid() bool {
	return it.node != nil
}

// Next advances the iterator to the next element of the map in insertion order
func (it *LinkedIterator) Next() {
	it.node = it.node.next
}

// Key returns the key of the underlying element
func (it *LinkedIterator) Key() interface{} {
	return it.node.k
}

// Value returns the value of the underlying element
func (it *LinkedIterator) Value() interface{} {
	return it.node.v
}

// ReverseLinkedIterator is used for iterating the LinkedOrderedMap in reverse insertion order.
type ReverseLinkedIterator struct {
	node *lrbtNode
}

// IsValid returns true if the iterator is valid for use, false otherwise.
// We must not call Next, Key, or Value if IsValid returns false.
func (it *ReverseLinkedIterator) IsValid() bool {
	return it.node != nil
}

// Next advances the iterator to the next element of the map in reverse insertion order
func (it *ReverseLinkedIterator) Next() {
	it.node = it.node.prev
}

// Key returns the key of the underlying element
func (it *ReverseLinkedIterator) Key() interface{} {
	return it.node.k
}

// Value returns the value of the underlying element
func (it *ReverseLinkedIterator) Value() interface{} {
	return it.node.v
}

type lrbtNodeType byte

const (
	kLRBTNodeTypeRoot lrbtNodeType = iota
	kLRBTNodeTypeLeftChild
	kLRBTNodeTypeRightChild
)

type lrbtNode struct {
	k           interface{}
	v           interface{}
	isBlack     bool
	nodeType    lrbtNodeType
	left        *lrbtNode
	right       *lrbtNode
	parent      *lrbtNode
	prev        *lrbtNode
	next        *lrbtNode
	orderedPrev *lrbtNode
	orderedNext *lrbtNode
}

func (node *lrbtNode) sibling() *lrbtNode {
	if node.parent != nil {
		if node.isLeftChild() {
			return node.parent.right
		}
		return node.parent.left
	}
	return nil
}

func (node *lrbtNode) rightmostChild() *lrbtNode {
	for node.right != nil {
		node = node.right
	}
	return node
}

func (node *lrbtNode) leftmostChild() *lrbtNode {
	for node.left != nil {
		node = node.left
	}
	return node
}

func (node *lrbtNode) isBlackNode() bool {
	if node != nil {
		return node.isBlack
	}
	return true
}

func (node *lrbtNode) isLeftChild() bool {
	return node.nodeType == kLRBTNodeTypeLeftChild
}

func (node *lrbtNode) isRightChild() bool {
	return node.nodeType == kLRBTNodeTypeRightChild
}
