#!/usr/bin/env bash
# Install the latest sshkeys binary and systemd units.
# Leave /etc/sshkeys and timer enablement unchanged.
set -euo pipefail

cleanup() {
    if [[ -n $publish ]]; then
        rm -f -- "$publish"
    fi
    rm -rf -- "$tmp"
}

install_file() {
    local mode=$1 source=$2 destination=$3

    # Stage beside the destination so replacement is one rename.
    publish=$(mktemp "${destination%/*}/.${destination##*/}.XXXXXX")
    install -m "$mode" "$source" "$publish"
    mv -fT -- "$publish" "$destination"
    publish=""
}

[[ $# -eq 0 ]] || { echo "Usage: sudo ./install.sh" >&2; exit 2; }
[[ $(id -u) -eq 0 ]] || { echo "Run as root." >&2; exit 1; }
[[ $(uname -s) == Linux ]] || { echo "Only Linux is supported." >&2; exit 1; }

case $(uname -m) in
    x86_64|amd64) arch=amd64 ;;
    aarch64|arm64) arch=arm64 ;;
    *) echo "Unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac

archive=sshkeys-linux-$arch.tar.gz
base=https://github.com/sumitgahoi/sshkeys/releases/latest/download
publish=""
tmp=$(mktemp -d)
trap cleanup EXIT

curl -fsSL --proto '=https' --proto-redir '=https' -o "$tmp/SHA256SUMS" "$base/SHA256SUMS"
curl -fsSL --proto '=https' --proto-redir '=https' -o "$tmp/$archive" "$base/$archive"

# GNU sha256sum text lines are "<hash><two spaces><name>".
awk -v file="$archive" '
    $2 == file && length($1) == 64 && $1 ~ /^[[:xdigit:]]+$/ {
        print $1 "  " file
        found = 1
    }
    END { exit !found }
' "$tmp/SHA256SUMS" > "$tmp/selected-sums" || {
    echo "Missing or invalid checksum for $archive" >&2
    exit 1
}
(cd "$tmp" && sha256sum --check --strict selected-sums)

tar -xzf "$tmp/$archive" -C "$tmp" sshkeys sshkeys-sync@.service sshkeys-sync@.timer

mkdir -p -m 0755 /usr/local/bin /etc/systemd/system
install_file 0755 "$tmp/sshkeys" /usr/local/bin/sshkeys
install_file 0644 "$tmp/sshkeys-sync@.service" /etc/systemd/system/sshkeys-sync@.service
install_file 0644 "$tmp/sshkeys-sync@.timer" /etc/systemd/system/sshkeys-sync@.timer

systemctl daemon-reload
echo "Installed /usr/local/bin/sshkeys and the systemd units."
