// run: sudo go run main.go run /bin/bash

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func run() {
	fmt.Println("Running", strings.Join(os.Args[2:], " "), fmt.Sprintf("as %d", os.Getpid()))
	// via /proc/self/exe, we create a new process running run() as well
	// what is different here is that we appended "child" in front of the arguments
	// for example, the original command is ./app run /bin/bash. if we run this cmd, the child command is
	// ./app child /bin/bash.
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...) // os.Args[1] == "run" / "child"
	// cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	// creating a new namespace requires sudo privileges
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		// CLONE_NEWUTS for creating a new Unix Time Sharing namespace (for hostname)
		// CLONE_NEWPID for creating a new PID namespace (doesn't affect `ps` which is attached to /proc *mounting*)
	}
	_ = cmd.Run()
}

func child() {
	fmt.Println("Child running", strings.Join(os.Args[2:], " "), fmt.Sprintf("as %d", os.Getpid()))

	// here the child process is in a new namespace.
	// we can do relative operations without affecting the host.
	// the reason why we create another child process is that before we run cmd in the child process, we are already inside
	// a new namespace, in which we can do some other initializations.
	// if we don't create another process, we will directly enter /bin/bash (for example) as soon as we create a new namespace,
	// leaving no space for further initialization.
	must(syscall.Sethostname([]byte("container")))
	// _ = syscall.Chroot("/vagrant/fs") // needs real bin/bash inside: /vagrant/fs/bin/bash => /bin/bash
	// syscall.Mount turns rootfs into a mount point, which is what pivot_root requires
	must(syscall.Mount("rootfs", "rootfs", "", syscall.MS_BIND, ""))
	// os.MkdirAll == mkdir -p
	must(os.MkdirAll("rootfs/oldroot", 0700))
	// the following line of code recursively sets the root mount and its children to private
	must(syscall.Mount("", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, ""))
	// pivot_root turns rootfs as the new root mount and put old root (and its sub mount tree) under
	// path_to_oldroot (rootfs/oldroot here)
	// after this, we may find rootfs under rootfs/oldroot (rootfs/oldroot/rootfs/oldroot/...)
	// that's not infinite loop; that's one mount relationship under path parsing
	// to avoid this, we detach the old root using syscall.Unmount, which is called afterward
	must(syscall.PivotRoot("rootfs", "rootfs/oldroot"))
	must(syscall.Chdir("/"))                              // change current work directory to cater for new root
	must(syscall.Unmount("/oldroot", syscall.MNT_DETACH)) // as mentioned before
	must(os.Remove("/oldroot"))                           // after detaching, remove the oldroot
	// /proc/<pid>/root -> /vagrant/fs (from ls -l)

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	// no need for namespace creating as we did it in the parent process
	_ = cmd.Run()
}

func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("bad command: " + strings.Join(os.Args, " "))
	}
}
