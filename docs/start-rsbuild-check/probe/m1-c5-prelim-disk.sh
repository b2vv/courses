#!/bin/zsh
# M1 (r3): repeat the two preliminary C5 builds (no-QEMU BuildKit image, m1-rosetta sanity build) with disk samples per build.
S=<scratch>
R=$S/r3; DF=<probe>/c5-buildkit-noqemu.Dockerfile
. <kiri>/environment.sh >/dev/null 2>&1
log() { print -r -- "$*" | tee -a $R/c5-prelim.log; }
disk() { docker exec courses-postgres-1 df -BG / | tail -1 | awk '{print $4}' | tr -d G; }
f=$(disk); log "START $(date -u +%FT%TZ) free=${f}G"; (( f < 10 )) && { log BLOCKED; exit 20; }
log "BUILD m1-buildkit-norqemu:local free_before=$(disk)G"
docker build -t m1-buildkit-norqemu:local -f $DF $R > $R/norqemu-build.log 2>&1
log "BUILD m1-buildkit-norqemu:local exit=$? id=$(docker image inspect m1-buildkit-norqemu:local --format '{{.Id}}') qemu_files=$(docker run --rm --entrypoint sh m1-buildkit-norqemu:local -c 'ls /usr/bin | grep -c qemu') free_after=$(disk)G"
docker buildx create --name m1-rosetta --driver docker-container --driver-opt image=m1-buildkit-norqemu:local --platform linux/amd64,linux/arm64 kiri > $R/builder-create.log 2>&1; log "BUILDER create exit=$?"
docker buildx inspect m1-rosetta --bootstrap > $R/builder-inspect.log 2>&1; log "BUILDER bootstrap exit=$? $(grep -E 'Platforms|BuildKit version' $R/builder-inspect.log | xargs)"
log "BUILD sanity free_before=$(disk)G"
docker buildx build --builder m1-rosetta --platform linux/amd64 --progress=plain --no-cache -f <probe>/c5-sanity.Dockerfile <probe> > $R/sanity-build.log 2>&1
log "BUILD sanity exit=$? out: $(grep -E '^#[0-9]+ [0-9.]+ (x86_64|aarch64|v24| 3e| b7)' $R/sanity-build.log | xargs) free_after=$(disk)G"
docker buildx rm m1-rosetta > /dev/null 2>&1; log "BUILDER rm exit=$? remaining=$(docker buildx ls 2>&1 | grep -c m1-rosetta)"
docker image rm m1-buildkit-norqemu:local > /dev/null 2>&1; log "IMAGE rm exit=$?"
log "END $(date -u +%FT%TZ) free=$(disk)G"
# Note: the recorded r3 run used an identical Dockerfile in the scratch dir $R/sanity as the build context.
