# wpool

A fork of [godoylucase/workers-pool](https://github.com/godoylucase/workers-pool) worker pool implementation with the modification of turing the execute function into a Task interface. Still need to work out the necessity of maintaining an Args parameter for a Job or just enforce custom types and handle Args as the type chooses. 

## Pool

The *pool* package is useful for sharing a static set of resource (e.g. database connections, memory buffers, etc). Goroutines can pull a resource from the pool, use it, and return it to the pool. *Pool* does not auto-reclaim resources after some timeout period. The calling program is responsible for ensuring any checked out resource is returned.

## Runner

The *runner* package is designed around executing a series of tasks within a certain time limit as well as gracefully exiting upon an `os.Interrupt` signal.
