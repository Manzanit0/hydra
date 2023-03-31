# Hydra

This is a small tool to build, test and lint services at scale in a monorepo.

## How to

You just have to run `hydra build` in any given directory and it will look for
Go module services to build:

```sh
$ hydra build 
🔎 looking for Go services...
👀 found 1 services
🏗  building service hydra
✅ hydra built succesfully in 387ms!
```

## FAQ

* What does `hydra` consider a Go service? Simple, opinionated: A Go module with a `main.go`.
* Where does it look for services? Recursively, in all subdirectories under the directory where you ran `hydra build`.
* How does it do concurrency? Naive: Spins up a goroutine for each build task, waits for them all.
