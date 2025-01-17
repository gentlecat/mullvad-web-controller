# Mullvad Web UI

I needed a lightweight web UI for easily switching between relay locations for a Mullvad client that's running on one of my servers. That server is proxying traffic for some of my devices. Occasionally, I need to quickly switch the current relay location, but SSH-ing to that server and making the switch with Mullvad CLI is a pain.

Note that I'm using this on **Debian** with **systemd**. If your setup is different, some changes might be needed.

## Installation

Having Mullvad with [the CLI](https://mullvad.net/en/help/how-use-mullvad-cli) installed is a prerequisite. You will
also need Go and NodeJS.

```shell
sudo apt install golang nodejs npm -y
```

Clone this repository. Build and install the systemd service:

```shell
make install
```

### Removal

```shell
make uninstall
```

## Usage

```shell
systemctl status mullvad-web-controller
```

URL to access the UI should be visible in the logs. By def
ault it runs on port 8666.
