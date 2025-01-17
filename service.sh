#!/bin/bash
set -e

install() {
  echo "Building and installing mullvad-web-controller service..."

  cp ./build/mullvad-web-controller /usr/local/bin/
  cp mullvad-web-controller.service /etc/systemd/system/

  systemctl daemon-reload
  systemctl enable mullvad-web-controller.service
  systemctl start mullvad-web-controller.service

  echo "mullvad-web-controller service installed and started."
}

uninstall() {
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
    install
    ;;
  uninstall)
    uninstall
    ;;
  *)
    echo "Invalid argument. Use 'install' or 'uninstall'."
    exit 1
    ;;
esac
