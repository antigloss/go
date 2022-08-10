# Overview

Package queue offers goroutine-safe Queue implementations such as LockfreeQueue(Lock free queue).

# LockfreeQueue

LockfreeQueue is a goroutine-safe Queue implementation. The overall performance of LockfreeQueue is much better than List+Mutex(standard package).

## Basic example

    lfq := queue.NewLockfreeQueue[int]() // create a LockfreeQueue
    lfq.Push(100) // Push an element into the queue
    v, ok := lfq.Pop() // Pop an element from the queue
