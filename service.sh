#!/bin/bash
set -e

install_service() {
  echo "Building and installing mullvad-web-controller service..."

  install -m 0755 ./build/mullvad-web-controller /usr/local/bin/
  install -m 0644 mullvad-web-controller.service /etc/systemd/system/

  systemctl daemon-reload
  systemctl enable mullvad-web-controller.service
  # Restart rather than start, so that reinstalling picks up the new binary.
  systemctl restart mullvad-web-controller.service

  echo "mullvad-web-controller service installed and started."
  echo "To change the settings, set environment variables in /etc/default/mullvad-web-controller and restart the service."
}

uninstall_service() {
  echo "Uninstalling mullvad-web-controller service..."

  systemctl stop mullvad-web-controller.service
  systemctl disable mullvad-web-controller.service

  rm /etc/systemd/system/mullvad-web-controller.service
  rm /usr/local/bin/mullvad-web-controller

  systemctl daemon-reload

  echo "mullvad-web-controller service uninstalled."
}

if [[ $# -eq 0 ]]; then
  echo "Usage: $0 install|uninstall"
  exit 1
fi

case "$1" in
  install)
    install_service
    ;;
  uninstall)
    uninstall_service
    ;;
  *)
    echo "Invalid argument. Use 'install' or 'uninstall'."
    exit 1
    ;;
esac
