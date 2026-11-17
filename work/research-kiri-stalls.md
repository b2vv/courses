# Kiri stand stalls: research (2026-11-17)

Question: why did `make check` attempts 4 and 5 on the Kiri stand fail with multi-second stalls?
- Attempt 4: W07/W08 got a 503 from identity after 4092 ms.
- Attempt 5: `catalog.spec.ts:2` got a 504 when SSR `/courses` took 14026 ms. The api answered its own call in 18 ms.

Every number below is read-only from the live stand. No setting was changed, and no gate was run.

## Facts

| # | Fact | Method |
|---|---|---|
| F1 | The Kiri VM is 6 vCPU / 16 GiB (`~/.kiri/vm.json`). Kiri 0.2.1 runs on Apple Virtualization.framework with `VZVirtioTraditionalMemoryBalloonDevice`. | `cat ~/.kiri/vm.json`; `~/.kiri/logs/daemon.log` "engine ready"; `strings /usr/local/bin/kiri` |
| F2 | Kiri's own UI text: memory is "the most the VM may take. It uses memory only as containers need it and gives the rest back to macOS". That is elastic memory through the balloon. Daemon knob `KIRI_NO_BALLOON` exists, next to `KIRI_MEMORY_MIB` and `KIRI_CPUS`. | `strings /usr/local/bin/kiri` |
| F3 | At 08:3x local, with the Courses stand idle: `nr_balloon_pages` = 3,018,240, so ≈11.5 GiB of the 16 GiB is ballooned away. The guest kernel had `MemAvailable` 1.7 GiB, page cache only 684 MiB, `AnonPages` 2.9 GiB. Counters: `allocstall_normal` 294,598; `pgmajfault` 19.2 M; `balloon_inflate` ≈ `balloon_deflate` ≈ 1.22 × 10⁹ pages cumulative. | `/proc/vmstat`, `/proc/meminfo` via `docker --context kiri exec courses-frontend-1` |
| F4 | **The balloon runs a sawtooth with a period of about 20 s, even when idle.** In 60 s sampled every 3 s, the balloon went 10 → 12 GiB, and guest `MemAvailable` fell to **31 / 139 / 146 MiB** at the troughs before Kiri deflated. Every trough coincides with a major-fault burst of **4,349–6,112 per 3 s**, against a baseline of 60–300. One 10 s window deflated 407,552 pages (1.55 GiB). | 20 samples of `nr_balloon_pages`, `MemAvailable`, `pgmajfault` (command in §Repeat) |
| F5 | Guest CPU and I/O pressure while Courses is idle: loadavg 5.9 on 6 CPUs. PSI values: cpu some avg10 8.6 %, memory full avg10 6.3 %, io full avg10 6.5 %. | `/proc/loadavg`, `/proc/pressure/*` |
| F6 | Co-tenants on the same VM. `factory-storage-alpha-factory-1` (AI Factory archive) uses **89–101 % CPU** continuously, and its postgres 0–56 %. There are also 8 cargotrans containers (4 MySQL/Maria DBs at ≈450 MiB each) and buildkit. None has a CPU limit (`NanoCpus=0`). | `docker --context kiri stats --no-stream` ×3; `inspect` |
| F7 | Host: 32 GiB RAM, swap **9.35 of 10 GiB used**, compressor holds ≈2.6 M pages. The Kiri VM process has 3.8–4.9 GiB RSS. No thermal or performance warning. The host did not sleep 07:56–08:23. | `sysctl vm.swapusage`, `vm_stat`, `ps`, `pmset -g log`, `pmset -g therm` |
| F8 | The 504 timeline (attempt 5). nginx received the request at 06:20:04.05. The frontend began it at 06:20:06.93, a **2.9 s accept delay**, and finished it at 06:20:20.95, after 14,026 ms. The api served its call at 06:20:06.96 in 18 ms. The next `/courses` took 53 ms. The frontend is native arm64: 0 restarts, not OOM-killed. Postgres is amd64 under emulation. `/readyz` took 423–771 ms, and once 3007 ms (503). | `frontend/test-logs/api-nginx-20261117T060428Z.log`; `docker compose logs frontend` |

## Diagnosis

**Mechanism (a reviewer can try it):** Kiri's balloon keeps shrinking the guest until it has almost no free memory left (F4). At each trough the kernel evicts file-backed pages: Node's and Postgres's code and data. The processes then block on major faults, here reads through virtio-blk, while the host is itself swapping (F7). That freezes a process for seconds, and both the 2.9 s accept delay and the 14 s render in F8 are single-process freezes on an otherwise idle request path. The co-tenant CPU load (F6) and the emulated amd64 postgres stretch every fault and every query.

**Not yet proven:** that the 06:20 stall fell on a trough. Nothing recorded balloon state during the gate. What is measured is the magnitude of the troughs and their period now. The 14 s stall itself is not traced to a trough.

**Not the cause:** the MS3c diff (catalog SSR is untouched); host sleep or thermal throttling (F7); OOM kills (`oom_kill 0`, frontend not OOM-killed).

## Options (each is the owner's decision)

1. **`KIRI_NO_BALLOON=1` for the Kiri daemon.** It is set on the LaunchAgent `dev.kiri.daemon` and needs a VM restart. The guest keeps its memory, so troughs and fault storms should stop. Risks:
   - the VM then pins up to 16 GiB on a host already at 9.35/10 GiB swap, so either lower `memory_mib` (for example 8–10 GiB) or free host memory;
   - it affects every project on Kiri.
2. **Remove co-tenant load from the gate window.** Stop or pause `factory-storage-alpha-*` and cargotrans during `make check`, or move them back to Docker Desktop. The README says the AI Factory archive default stays `desktop-linux`, yet it runs on Kiri (F6).
3. **Native arm64 postgres image.** It removes the emulation tax. Per the Kiri README it is a separate decision: it touches the cut-over data volume.
4. **Run the next gate with a recorder:** sample `nr_balloon_pages`, `MemAvailable` and `pgmajfault` every 2 s into a log. This lets the next stall be correlated or excluded. It is cheap, read-only, and can be combined with 1–3.

## Repeat

```sh
for i in $(seq 1 20); do docker --context kiri exec courses-frontend-1 sh -c \
  'echo "$(date +%T) balloon_pages=$(awk "/^nr_balloon_pages /{print \$2}" /proc/vmstat) \
   avail_kB=$(awk "/MemAvailable/{print \$2}" /proc/meminfo) \
   majflt=$(awk "/^pgmajfault /{print \$2}" /proc/vmstat)"'; sleep 3; done
```

Sources: none public; a web search for Kiri's memory model found nothing. Everything comes from the local binary, logs and `/proc`.

## Update 2026-11-17: cargotrans stopped (owner: "cargotrans туши")

The 8 `cargotrans-server-api` containers were stopped with `docker --context kiri stop`. Their restart policy is `unless-stopped`, so
they stay down. Containers and volumes are kept; DBs exit 0, apps 137 after the 10 s grace. Re-sampled 30 s every 3 s:
- loadavg fell from 5.9 to **4.3–5.0**;
- the balloon simply took the freed memory: **12→14 GiB** (was 10→12);
- the troughs are unchanged or deeper: `MemAvailable` **17 / 59 MiB**, with a 4,200 major-fault burst in 3 s;
- host swap is 9.16 of 10 GiB.

**Conclusion:** removing co-tenants lowers CPU load but does not stop the memory troughs. Kiri's balloon steers the guest to
near-zero free memory whatever the load (F4). Only `KIRI_NO_BALLOON` (option 1) addresses that mechanism.

## Update 2026-11-17: balloon disabled (owner: "Так вимикай")

Applied option 1.
- **Change:** `EnvironmentVariables.KIRI_NO_BALLOON=1` added to `~/Library/LaunchAgents/dev.kiri.daemon.plist`. The backup is
  `/Users/strelia/project-data/courses-kiri/dev.kiri.daemon.plist.bak-20261117-before-noballoon`.
- **Restart:** `launchctl bootout` then `bootstrap`; the daemon was gone after 6 s. "VM booted" was logged at 07:12:12Z, still at
  6 CPU / 16384 MiB (not lowered).
- **Effect:** `nr_balloon_pages` 0 and `balloon_inflate` 0 for 90 s after boot. Guest `MemAvailable` holds 14.9–15.0 GiB, major
  faults run at about 50–200 per 3 s (troughs used to reach 4–6 k), loadavg is 1.1–1.5. The virtio_balloon driver is still bound but
  stays idle.
- **Stand:** all 10 Courses services and the 3 factory-storage services came back healthy. `courses-api` restarted 3× on the engine
  restart: the ordering burst the Kiri README documents. `GET /courses` returns 200 in 0.28 s.
- **Watch:** the VM is no longer returning memory to macOS. Its RSS was 4.07 GiB just after boot; host swap was 9.13 of 10 GiB. As
  guest page cache grows, host pressure rises. If the host starts thrashing, lower `memory_mib` to 8–10 GiB, which is the owner's call.
- **Rollback:** `cp -p <backup> ~/Library/LaunchAgents/dev.kiri.daemon.plist`, then
  `launchctl bootout gui/$(id -u)/dev.kiri.daemon; launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/dev.kiri.daemon.plist`.
  A Kiri update or `kiri setup` may rewrite the plist; re-check
  `launchctl print gui/$(id -u)/dev.kiri.daemon | grep KIRI_NO_BALLOON` after either.
- **Not proven yet:** that this removes the gate's 503/504 stalls. That needs a gate run with the balloon/fault recorder.

## Update 2026-11-17: first gate with the balloon off (attempt 6)

Recorder: 587 samples over 20 min, one every 2 s. Balloon pages 0, guest `MemAvailable` min 11.1 GiB, `allocstall_normal` 0. The
largest burst was 4,864 major faults per 2 s at 07:22:19, during the image build and start. Peak loadavg was 9.6. The api/nginx stream
had **0 responses with 5xx**. Attempt 5 had five: three 503s, one of them `/readyz` after 3 s, and a 504 after 15 s. **Conclusion:** with
the balloon off, the stall class did not recur in a full gate. That is one run, so it is correlation backed by the measured mechanism,
not proof of absence.

A second Kiri cut-over defect surfaced in the same run: the copied volumes carry no compose labels, so `make e2e-reset`, which selects
by label, silently removes nothing. See `docs/multi-subtitles/report.md`, attempt 6.
