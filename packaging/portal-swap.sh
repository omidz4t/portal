#!/bin/bash
# Atomic replace of /usr/bin/portal. Reads the new ELF from stdin.
# Used as SSH forced-command for the GitHub Actions deploy key.
set -euo pipefail

BIN=/usr/bin/portal
NEW=/usr/bin/portal.new
PREV=/usr/bin/portal.prev
UNIT=portal
if ! systemctl list-unit-files portal.service >/dev/null 2>&1; then
	if systemctl list-unit-files tgportal.service >/dev/null 2>&1; then
		UNIT=tgportal
	fi
fi
STAGED=$(mktemp /tmp/portal.XXXXXX)
trap 'rm -f "$STAGED"' EXIT

umask 077
cat >"$STAGED"

if ! file -b "$STAGED" | grep -q 'ELF 64-bit'; then
	echo "error: stdin is not a 64-bit ELF binary" >&2
	exit 1
fi
chmod 0755 "$STAGED"

if ! "$STAGED" --version >/dev/null 2>&1; then
	echo "error: new binary failed --version" >&2
	exit 1
fi
VER=$("$STAGED" --version | head -1 | tr -d '\r')
echo "staging ${VER}"

sudo install -m 0755 -o root -g root "$STAGED" "$NEW"
if [[ -x "$BIN" ]]; then
	sudo mv -f "$BIN" "$PREV"
fi
sudo mv -f "$NEW" "$BIN"
# Keep old name working until packages catch up.
if [[ ! -e /usr/bin/tgportal ]]; then
	sudo ln -s portal /usr/bin/tgportal
fi
sudo systemctl restart "$UNIT"

ok=0
for _ in $(seq 1 20); do
	if systemctl is-active --quiet "$UNIT"; then
		ok=1
		break
	fi
	sleep 1
done

if [[ "$ok" -ne 1 ]]; then
	echo "error: service failed to become active — rolling back" >&2
	if [[ -x "$PREV" ]]; then
		sudo mv -f "$PREV" "$BIN"
		sudo systemctl restart "$UNIT" || true
	fi
	systemctl --no-pager --full status "$UNIT" || true
	exit 1
fi

echo "deployed ${VER} unit=${UNIT} pid=$(systemctl show -p MainPID --value "$UNIT")"
