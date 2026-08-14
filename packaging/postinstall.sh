#!/bin/sh
set -e
if command -v systemd-sysusers >/dev/null 2>&1; then
	systemd-sysusers tgportal.conf || true
fi
if command -v systemd-tmpfiles >/dev/null 2>&1; then
	systemd-tmpfiles --create /usr/lib/tmpfiles.d/tgportal.conf || true
fi
if [ ! -f /etc/tgportal/config.yml ] && [ -f /usr/share/tgportal/config.example.yml ]; then
	mkdir -p /etc/tgportal
	cp /usr/share/tgportal/config.example.yml /etc/tgportal/config.yml
	chmod 0640 /etc/tgportal/config.yml
	chown root:tgportal /etc/tgportal/config.yml 2>/dev/null || true
fi
if command -v systemctl >/dev/null 2>&1; then
	systemctl daemon-reload || true
fi
echo "tgportal installed. Copy secrets to /etc/tgportal/env and: systemctl enable --now tgportal"
