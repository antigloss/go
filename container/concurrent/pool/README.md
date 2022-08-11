# Overview

Package pool offers some goroutine-safe object pools such as ObjectPool.

# ObjectPool

ObjectPool is a goroutine-safe generic pool for objects of any type.

## Basic example

	// create an ObjectPool for bytes.Buffer
	op := pool.NewObjectPool[bytes.Buffer](10000,
	                         func() *bytes.Buffer { return new(bytes.Buffer) },
	                         func(obj *bytes.Buffer) { obj.Reset() })
	obj := op.Get() // get a ready-to-use bytes.Buffer
	// do something with `buf`
	op.Put(obj) // return obj to ObjectPool.
