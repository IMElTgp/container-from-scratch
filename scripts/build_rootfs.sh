#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
rootfs="${repo_root}/rootfs"

binaries=(
  /usr/bin/bash
  /usr/bin/ls
  /usr/bin/cat
  /usr/bin/echo
  /usr/bin/env
  /usr/bin/hostname
  /usr/bin/mkdir
  /usr/bin/mount
  /usr/bin/ps
  /usr/bin/pwd
  /usr/bin/rm
  /usr/bin/sleep
  /usr/bin/touch
  /usr/bin/uname
)

copy_binary() {
  local src="$1"
  local dest="${rootfs}${src}"

  if [[ -e "${dest}" ]]; then
    return
  fi

  install -D -m 0755 "${src}" "${dest}"
}

copy_library() {
  local src="$1"
  local dest="${rootfs}${src}"

  if [[ -e "${dest}" ]]; then
    return
  fi

  install -D -m 0755 "${src}" "${dest}"
}

copy_dependencies() {
  local bin="$1"
  local dep

  while IFS= read -r dep; do
    [[ -n "${dep}" ]] || continue
    copy_library "${dep}"
  done < <(ldd "${bin}" | awk '/=> \// { print $3 } $1 ~ /^\// { print $1 }')
}

rm -rf "${rootfs}"
install -d -m 0755 "${rootfs}/usr/bin" "${rootfs}/usr/lib" "${rootfs}/usr/lib64"
install -d -m 0755 "${rootfs}/etc" "${rootfs}/dev" "${rootfs}/proc" "${rootfs}/root"
install -d -m 1777 "${rootfs}/tmp"
ln -s usr/bin "${rootfs}/bin"
ln -s usr/lib "${rootfs}/lib"
ln -s usr/lib64 "${rootfs}/lib64"

printf 'root:x:0:0:root:/root:/bin/bash\n' > "${rootfs}/etc/passwd"
printf 'root:x:0:\n' > "${rootfs}/etc/group"
printf 'container\n' > "${rootfs}/etc/hostname"

for bin in "${binaries[@]}"; do
  copy_binary "${bin}"
  copy_dependencies "${bin}"
done

ln -sf bash "${rootfs}/usr/bin/sh"

printf 'Built rootfs at %s\n' "${rootfs}"
