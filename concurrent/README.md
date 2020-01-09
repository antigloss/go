# Overview

Package concurrent provides some concurrent control utilities and goroutine-safe containers.

## Semaphore

### Usage

``` 
// Creates a ready-to-use semaphare
sema := concurrent.NewSemaphore(InitValue)
// Decrements the semaphore, blocks if the semaphore is less than 1
sema.Acquire()
// Increments the semaphore. If the semaphore¡¯s value consequently becomes greater than zero,
// then another goroutine blocked in sema.Acquire() will be woken up and proceed to lock the semaphore.
sema.Release()
```
