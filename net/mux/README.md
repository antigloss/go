# Overview

Package mux (short for connection multiplexer) is a multiplexing package for Golang.

In some rare cases, we can only open a few connections to a remote server, but we want to program like we can open unlimited connections. Should you encounter this rare cases, then this package is exactly what you need.

# SimpleMux

SimpleMux is a connection multiplexer. It is very useful when under the following constraints:

  1. Can only open a few connections (probably only 1 connection) to a remote server,
     but want to program like there can be unlimited connections.
  2. The remote server has its own protocol format which could not be changed.
  3. Fortunately, we can set 8 bytes of information to the protocol header which
     will remain the same in the server's response.

## Basic example

Seek to simple_mux_test.go for detailed usage.
