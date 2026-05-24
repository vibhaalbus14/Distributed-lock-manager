# 🛡️ Distributed Lock Manager with Network Simulation Dashboard



An interactive, real-time distributed systems simulation engine demonstrating thread-safe mutual exclusion, FIFO queue management, heartbeats, lease management, and packet manipulation under network strain.

## 🛠️ System Features
* **Distributed Mutual Exclusion:** Coordinates atomic lock leases using a centralized manager pattern.
* **Fencing Tokens ($Token_{gen}$):** Implements monotonically increasing sequence IDs to neutralize late-arriving or delayed network packets.
* **Lease Time-To-Live (TTL):** Incorporates OS-level timers (`time.Timer`) to track lock expirations if a node encounters network isolation or a hard failure.
* **Dynamic Network Simulator:** Empowers operators to introduce packet-drop risk and execution latencies dynamically.
* **Zero-Downtime Soft-Reset:** Purges node registries and cluster variables instantly via synchronized mutex controls while keeping core memory channels awake and operational.
...

  Live link: https://distributed-lock-manager-simulator.netlify.app/
