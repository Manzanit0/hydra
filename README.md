# Hydra

This is a small tool to build, test and lint services at scale in a monorepo.

## How to

You just have to run `hydra build` in any given directory and it will look for
Go module services to build:

```sh
$ hydra build --owners @Manzanit0 --names hydra
🔎 looking for Go services...
👀 found 1 services
🍕 filtering services with names: hydra
🔑 filtering services owned by: @Manzanit0
🏗  building service hydra
✅ hydra built succesfully in 354ms!
```

In case you want to run tests instead of building:

```sh
$ hydra test
🔎 looking for Go services...
👀 found 1 services
✅ hydra test succesfully in 7250ms!
?   	github.com/manzanit0/hydra	[no test files]
?   	github.com/manzanit0/hydra/pkg/owner	[no test files]
?   	github.com/manzanit0/hydra/pkg/tool	[no test files]
```

## FAQ

* What does `hydra` consider a Go service? Simple, opinionated: A Go module with a `main.go`.
* Where does it look for services? Recursively, in all subdirectories under the directory where you ran `hydra build`.
* How does it do concurrency? Naive: Spins up a goroutine for each build task, waits for them all.
