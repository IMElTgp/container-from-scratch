# Container From Scratch

In this repository, I implemented an MVP container, in which I created an isolated environment:

- PID, hostname and mount isolation
  - `PID namespace` as `syscall.CLONE_NEWPID`
  - `UTS namespace` as `syscall.CLONE_NEWUTS`
  - `mount namespace` as `syscall.CLONE_NEWNS`
- New root mount via `syscall.PivotRoot`
- Basic rootfs with `/bin/bash`

## Quick Start

```bash
bash scripts/build_rootfs.sh
# with root privileges:
go run main.go run /bin/bash
# or, without root privileges:
sudo go run main.go run /bin/bash
```

## References 

- [Build Your Own Container Using Less than 100 Lines of Go](https://www.infoq.com/articles/build-a-container-golang/)
- [Containers From Scratch • Liz Rice • GOTO 2018](https://www.youtube.com/watch?v=8fi7uSYlOdc)