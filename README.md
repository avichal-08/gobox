# GoBox

A daemonless container runtime built from scratch in Go. GoBox demystifies how tools like Docker work under the hood by directly interacting with Linux kernel primitives.

---

## Overview

Modern container runtimes are often treated as black boxes. GoBox tears that abstraction away. Written entirely in Go, it implements the core infrastructure of containerization by working directly with the Linux kernel: namespaces, cgroups, `pivot_root`, and OverlayFS.

The result is a fully functional, educational runtime that creates genuine process isolation — the same fundamental mechanisms that power production-grade container infrastructure.

---

## Features

**Process and Identity Isolation**
Uses Linux `clone` system calls (`CLONE_NEWPID`, `CLONE_NEWUTS`, `CLONE_NEWNS`) to isolate process trees, hostnames, and mount points. Each container gets its own PID namespace, meaning processes inside cannot see or signal processes on the host.

**True Filesystem Sandboxing**
Secures the container by executing `pivot_root` at the kernel level, physically swapping the root filesystem rather than relying on `chroot`. This eliminates the well-documented escape vulnerabilities inherent to `chroot`-based sandboxing.

**Ephemeral Layered Filesystem**
Implements Docker-style layered storage via OverlayFS. Containers write to an isolated, temporary `upperdir` layer that is discarded on exit, leaving the read-only `lowerdir` base image completely untouched. Each run starts from a clean slate.

**Resource Throttling via Cgroups v2**
Protects the host from container resource exhaustion — including fork bombs — by dynamically writing resource limits directly to the Linux `/sys/fs/cgroup` virtual filesystem. CPU and memory constraints are enforced at the kernel level.

---

## Architecture

GoBox follows a parent-child execution model that mirrors the lifecycle of a real container runtime.

```
gobox run <cmd>
      |
      v
 [ Parent Process ]
      |
      |-- Requests isolated namespaces from the Linux kernel (CLONE_NEWPID, CLONE_NEWUTS, CLONE_NEWNS)
      |-- Configures OverlayFS: stacks read-only Alpine base image with a temporary writable layer
      |-- Writes cgroup resource limits to /sys/fs/cgroup
      |
      v
 [ Filesystem Engine ]
      |
      |-- Mounts the OverlayFS union at a staging path
      |-- Executes pivot_root, atomically swapping / to the container's filesystem
      |-- Unmounts and detaches all host filesystem references
      |
      v
 [ Child Process ]
      |
      |-- Wakes up as PID 1 inside a fully isolated environment
      |-- Executes the user payload (e.g., /bin/sh)
      |-- On exit, the writable upperdir layer is discarded entirely
```

### Lifecycle Summary

| Stage | Component | Responsibility |
|---|---|---|
| 1 | CLI (`gobox run`) | Forks itself, requests new kernel namespaces |
| 2 | Filesystem Engine | Assembles the OverlayFS union mount |
| 3 | Pivot | Swaps root via `pivot_root`, detaches host mounts |
| 4 | Child | Runs as PID 1 inside the isolated container |

---

## Prerequisites

GoBox uses Linux kernel primitives directly and therefore requires:

- A Linux environment (native or WSL2)
- Root privileges for namespace and mount operations
- Go 1.21 or later
- An Alpine Linux root filesystem (see setup below)

---

## Getting Started

**1. Clone the repository**

```bash
git clone https://github.com/yourusername/gobox.git
cd gobox
```

**2. Prepare a base Alpine root filesystem**

GoBox expects an unpacked Alpine Linux rootfs in a directory named `alpine-fs`. The simplest way to obtain one is via Docker:

```bash
docker export $(docker create alpine) | tar -C alpine-fs -xvf -
```

**3. Build the GoBox binary**

```bash
go build -o gobox .
```

**4. Run an isolated shell**

```bash
sudo ./gobox run sh
```

You are now inside a containerized environment with an isolated PID namespace, hostname, filesystem, and resource limits. The host machine remains completely unaffected.

---

## Implementation Notes

**Why `pivot_root` instead of `chroot`?**
`chroot` changes the apparent root directory for a process but does not prevent a privileged process from escaping it. `pivot_root` is a kernel-level operation that physically reassigns the root mount point of a process's mount namespace, making escape structurally impossible rather than merely inconvenient.

**Why Cgroups v2?**
The unified cgroup hierarchy introduced in v2 simplifies resource management significantly. GoBox writes directly to the `/sys/fs/cgroup` virtual filesystem rather than using a library, keeping the implementation transparent and auditable.

**Why OverlayFS?**
OverlayFS enables Docker-style image layering without copying data. The base image (lowerdir) is never written to. Containers accumulate changes in a fast, copy-on-write upper layer that disappears on exit — making container startup nearly instantaneous.

---

## What I Learned

Building GoBox fundamentally changed how I think about modern infrastructure.

The central insight is that containers are not virtual machines. There is no hypervisor, no emulated hardware, no guest OS. A container is an ordinary Linux process with a carefully constructed set of kernel-level illusions — a different view of the process tree, a different root filesystem, a different hostname, resource limits enforced by the scheduler.

The practical takeaways:

- **Low-level Go systems programming** — working directly with `syscall.SysProcAttr`, clone flags, and mount syscalls rather than relying on higher-level abstractions.
- **Linux filesystem internals** — understanding the difference between bind mounts, OverlayFS union mounts, and mount namespaces, and why the order of operations matters.
- **OS-level security primitives** — learning why `chroot` is insufficient for true sandboxing and how namespace isolation actually works at the kernel level.
- **Cgroup accounting** — understanding how the Linux scheduler enforces resource constraints and why writing raw values to the cgroup filesystem is equivalent to what `docker run --memory` does internally.

The project is intentionally kept simple and readable. Every design decision prioritizes clarity over performance, making it a useful reference for anyone who wants to understand what happens below the Docker API.

---

## License

MIT