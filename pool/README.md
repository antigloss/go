# Overview

First of all, sorry for my poor English, I've tried my best.

Package pool offers goroutine-safe object pools such as ObjectPool and BufferPool.

ObjectPool is a generic object pool that can be used to pool objects of any type, it performs about 36% slower than the other specialized object pools such as BufferPool.

# ObjectPool

ObjectPool is a goroutine-safe generic pool for objects of any type. It performs about 36% slower than the other specialized object pools such as BufferPool.

## Basic example

  // create an ObjectPool for bytes.Buffer
  op := pool.NewObjectPool(10000,
                            func() interface{} { return new(bytes.Buffer) },
                            func(obj interface{}) { obj.(*bytes.Buffer).Reset() })
  obj := op.Get()
  buf := obj.(*bytes.Buffer)
  // do something with `buf`
  op.Put(obj)

# BufferPool

BufferPool is a goroutine-safe pool for bytes.Buffer.

## Basic example

  bp := pool.NewBufferPool(1000, 512)
  buf := bp.Get() // get a ready-to-use bytes.Buffer
  // do something with `buf`
  bp.Put(buf) // return buf to BufferPool
