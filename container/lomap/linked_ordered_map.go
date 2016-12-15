// Author: https://github.com/antigloss

// Package lomap implements an ordered map which supports iteration in insertion order.
// It's also optimized for ordered traverse. lomap is short for Linked Ordered Map.
//
// Caution: This package is not goroutine-safe!
package lomap

// Comparator compares a and b and returns:
//     0 if they are equal
//     < 0 if a < b
//     > 0 if a > b
type Comparator func(a, b interface{}) int

// LinkedOrderedMap is an ordered map which supports iteration in insertion order.
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

// Set inserts a new node into the LinkedOrderedMap or updates the existing node with the new value.
// Key should adhere to the comparator's type assertion, otherwise it will panic.
//   key: key of the value to be inserted/updated
//   value: value to be inserted/updated
func (t *LinkedOrderedMap) Set(key interface{}, value interface{}) {
	newNode := &lrbtNode{k: key, v: value}
	if t.root != nil {
		node := t.root
		for {
			ret := t.comp(key, node.k)
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
				node.k = key
				node.v = value
				return
			}
		}
		newNode.parent = node
		t.insertCase2(newNode)
		// insert ordered linked list
		newNode.prev = t.tail
		t.tail.next = newNode
		t.tail = newNode
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
				t.orderedHead = newNode
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
				t.orderedTail = newNode
			}
		} else {
			newNode.orderedPrev = newNode.left
			newNode.orderedNext = newNode.right
			newNode.left.orderedNext = newNode
			newNode.right.orderedPrev = newNode
		}
	} else {
		t.root = newNode
		t.head = newNode
		t.tail = newNode
		t.orderedHead = newNode
		t.orderedTail = newNode
		newNode.isBlack = true
		newNode.nodeType = kLRBTNodeTypeRoot
	}

	t.size++
}

// Get returns value of the key and true if the given key is found.
// If the given key is not found, it returns nil, false
// Key should adhere to the comparator's type assertion, otherwise it will panic.
func (t *LinkedOrderedMap) Get(key interface{}) ( /*value*/ interface{} /*found*/, bool) {
	node := t.search(key)
	if node != nil {
		return node.v, true
	}
	return nil, false
}

// Remove removes the node with the given key from the map.
// Key should adhere to the comparator's type assertion, otherwise it will panic.
func (t *LinkedOrderedMap) Remove(key interface{}) {
	node := t.search(key)
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
				t.tail = node.prev
			}
		}
		if node.next != nil && node.next != predecessor && node.prev != predecessor {
			node.next.prev = node.prev
			if node.prev == nil {
				t.head = node.next
			}
		}
		if predecessor.prev != node {
			node.prev = predecessor.prev
			if predecessor.prev != nil {
				predecessor.prev.next = node
			} else {
				t.head = node
			}
		}
		if predecessor.next != node {
			node.next = predecessor.next
			if predecessor.next != nil {
				predecessor.next.prev = node
			} else {
				t.tail = node
			}
		}

		// Fix ordered linked list
		if node.orderedPrev != nil && node.orderedPrev != predecessor && node.orderedNext != predecessor {
			node.orderedPrev.orderedNext = node.orderedNext
			if node.orderedNext == nil {
				t.orderedTail = node.orderedPrev
			}
		}
		if node.orderedNext != nil && node.orderedNext != predecessor && node.orderedPrev != predecessor {
			node.orderedNext.orderedPrev = node.orderedPrev
			if node.orderedPrev == nil {
				t.orderedHead = node.orderedNext
			}
		}
		if predecessor.orderedPrev != node {
			node.orderedPrev = predecessor.orderedPrev
			if predecessor.orderedPrev != nil {
				predecessor.orderedPrev.orderedNext = node
			} else {
				t.orderedHead = node
			}
		}
		if predecessor.orderedNext != node {
			node.orderedNext = predecessor.orderedNext
			if predecessor.orderedNext != nil {
				predecessor.orderedNext.orderedPrev = node
			} else {
				t.orderedTail = node
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
		t.deleteCase1(node)
	}
	t.replaceNode(node, child)
	// If the node that was deleted is a root node
	if node.parent == nil && child != nil {
		child.isBlack = true
	}

	if needFixList {
		// Fix insert ordered linked list
		if node.prev != nil {
			node.prev.next = node.next
		} else {
			t.head = node.next
		}
		if node.next != nil {
			node.next.prev = node.prev
		} else {
			t.tail = node.prev
		}

		// Fix ordered linked list
		if node.orderedPrev != nil {
			node.orderedPrev.orderedNext = node.orderedNext
		} else {
			t.orderedHead = node.orderedNext
		}
		if node.orderedNext != nil {
			node.orderedNext.orderedPrev = node.orderedPrev
		} else {
			t.orderedTail = node.orderedPrev
		}
	}

	t.size--
}

// Empty returns true if the map does not contain any nodes, otherwise it returns false.
func (t *LinkedOrderedMap) Empty() bool {
	return t.size == 0
}

// Size returns the number of nodes in the map.
func (t *LinkedOrderedMap) Size() int {
	return t.size
}

// Clear removes all nodes from the map.
func (t *LinkedOrderedMap) Clear() {
	t.root = nil
	t.head = nil
	t.tail = nil
	t.orderedHead = nil
	t.orderedTail = nil
	t.size = 0
}

// Count returns the number of elements with key key, which is either 1 or 0 since this container does not allow duplicates.
//
//   key: key value of the elements to count
func (t *LinkedOrderedMap) Count(key interface{}) int {
	if t.search(key) != nil {
		return 1
	}
	return 0
}

// Case 1: root node
func (t *LinkedOrderedMap) insertCase1(node *lrbtNode) {
	if node.parent != nil {
		t.insertCase2(node)
	} else { // Root node
		node.isBlack = true
	}
}

// Case 2: black node can have children of any color
func (t *LinkedOrderedMap) insertCase2(node *lrbtNode) {
	if !node.parent.isBlack {
		t.insertCase3(node)
	}
}

// Case 3: red nodes' children must be black
func (t *LinkedOrderedMap) insertCase3(node *lrbtNode) {
	uncle := node.parent.sibling()
	if !uncle.isBlackNode() {
		node.parent.isBlack = true
		uncle.isBlack = true
		node.parent.parent.isBlack = false
		t.insertCase1(node.parent.parent)
	} else {
		t.insertCase4(node)
	}
}

// Case 4
func (t *LinkedOrderedMap) insertCase4(node *lrbtNode) {
	if node.isRightChild() && node.parent.isLeftChild() {
		t.rotateLeft(node.parent)
		node = node.left
	} else if node.isLeftChild() && node.parent.isRightChild() {
		t.rotateRight(node.parent)
		node = node.right
	}
	t.insertCase5(node)
}

// Case 5
func (t *LinkedOrderedMap) insertCase5(node *lrbtNode) {
	node.parent.isBlack = true
	grandparent := node.parent.parent
	grandparent.isBlack = false
	if node.isLeftChild() && node.parent.isLeftChild() {
		t.rotateRight(grandparent)
	} else if node.isRightChild() && node.parent.isRightChild() {
		t.rotateLeft(grandparent)
	}
}

// Case 1: root node
func (t *LinkedOrderedMap) deleteCase1(node *lrbtNode) {
	if node.parent != nil {
		t.deleteCase2(node)
	}
}

// Case 2: sibling node is red
func (t *LinkedOrderedMap) deleteCase2(node *lrbtNode) {
	sibling := node.sibling()
	if !sibling.isBlackNode() {
		node.parent.isBlack = false
		sibling.isBlack = true
		if node.isLeftChild() {
			t.rotateLeft(node.parent)
		} else {
			t.rotateRight(node.parent)
		}
	}
	t.deleteCase3(node)
}

// Case 3: parent, sibling and its children are black
func (t *LinkedOrderedMap) deleteCase3(node *lrbtNode) {
	sibling := node.sibling()
	if node.parent.isBlack && sibling.isBlack && sibling.left.isBlackNode() && sibling.right.isBlackNode() {
		sibling.isBlack = false
		t.deleteCase1(node.parent)
	} else {
		t.deleteCase4(node, sibling)
	}
}

// Case 4: parent is red and sibling and its children are black
func (t *LinkedOrderedMap) deleteCase4(node, sibling *lrbtNode) {
	if !node.parent.isBlack && sibling.isBlack && sibling.left.isBlackNode() && sibling.right.isBlackNode() {
		sibling.isBlack = false
		node.parent.isBlack = true
	} else {
		t.deleteCase5(node, sibling)
	}
}

// Case 5: only one child of sibling is red
func (t *LinkedOrderedMap) deleteCase5(node, sibling *lrbtNode) {
	if node.isLeftChild() && sibling.isBlack && !sibling.left.isBlackNode() && sibling.right.isBlackNode() {
		sibling.isBlack = false
		sibling.left.isBlack = true
		t.rotateRight(sibling)
	} else if node.isRightChild() && sibling.isBlack && !sibling.right.isBlackNode() && sibling.left.isBlackNode() {
		sibling.isBlack = false
		sibling.right.isBlack = true
		t.rotateLeft(sibling)
	}
	t.deleteCase6(node)
}

// Case 6
func (t *LinkedOrderedMap) deleteCase6(node *lrbtNode) {
	sibling := node.sibling()
	sibling.isBlack = node.parent.isBlack
	node.parent.isBlack = true
	if node.isLeftChild() && !sibling.right.isBlackNode() {
		sibling.right.isBlack = true
		t.rotateLeft(node.parent)
	} else if !sibling.left.isBlackNode() {
		sibling.left.isBlack = true
		t.rotateRight(node.parent)
	}
}

func (t *LinkedOrderedMap) rotateLeft(node *lrbtNode) {
	right := node.right
	t.replaceNode(node, right)
	node.right = right.left
	if right.left != nil {
		right.left.parent = node
		right.left.nodeType = kLRBTNodeTypeRightChild
	}
	right.left = node
	node.parent = right
	node.nodeType = kLRBTNodeTypeLeftChild
}

func (t *LinkedOrderedMap) rotateRight(node *lrbtNode) {
	left := node.left
	t.replaceNode(node, left)
	node.left = left.right
	if left.right != nil {
		left.right.parent = node
		left.right.nodeType = kLRBTNodeTypeLeftChild
	}
	left.right = node
	node.parent = left
	node.nodeType = kLRBTNodeTypeRightChild
}

func (t *LinkedOrderedMap) search(key interface{}) (node *lrbtNode) {
	node = t.root
	for node != nil {
		ret := t.comp(key, node.k)
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

func (t *LinkedOrderedMap) replaceNode(oldNode *lrbtNode, newNode *lrbtNode) {
	if oldNode.parent == nil {
		t.root = newNode
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
