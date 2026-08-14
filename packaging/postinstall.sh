#!/bin/sh
set -e
if command -v systemd-sysusers >/dev/null 2>&1; then
	systemd-sysusers portal.conf || true
fi
if command -v systemd-tmpfiles >/dev/null 2>&1; then
	systemd-tmpfiles --create /usr/lib/tmpfiles.d/portal.conf || true
fi
if [ ! -f /etc/portal/config.yml ] && [ -f /usr/share/portal/config.example.yml ]; then
	mkdir -p /etc/portal
	cp /usr/share/portal/config.example.yml /etc/portal/config.yml
	chmod 0640 /etc/portal/config.yml
	chown root:portal /etc/portal/config.yml 2>/dev/null || true
fi
if command -v systemctl >/dev/null 2>&1; then
	systemctl daemon-reload || true
fi
echo "portal installed. Copy secrets to /etc/portal/env and: systemctl enable --now portal"
