#!/bin/bash

if [[ -z "$(which crudini)" ]]; then
    echo "crudini not found"
    exit 1
fi

PROXY_FULL_URL="${1}"

SYSTEMD_FILE="/etc/systemd/system/docker.service.d/http-proxy.conf"
if [[ ! -e "$(dirname "${SYSTEMD_FILE}")" ]]; then
    mkdir "$(dirname "${SYSTEMD_FILE}")" || exit 1
fi

if [[ -z "${PROXY_FULL_URL}" ]]; then
    crudini --del "${SYSTEMD_FILE}" "Service" "Environment" || exit 1
else
    crudini --set "${SYSTEMD_FILE}" "Service" "Environment" "HTTP_PROXY=${PROXY_FULL_URL}" || exit 1
fi

systemctl daemon-reload || exit 1
systemctl restart docker.service || exit 1
