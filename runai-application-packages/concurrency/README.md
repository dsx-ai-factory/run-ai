# concurrency

Directory for resources and tools for implementing concurrent access to data.

## ConcurrentMap
`ConcurrentMap` is a generic type on top of `xsync.Map`. 

in the code we use this map extensively, as it allows
read/write with no locking, but its value is defined as 
`interface{}` which is not comfortable to work with,
that's why we use this generic type.
