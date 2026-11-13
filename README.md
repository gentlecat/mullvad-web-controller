# Mullvad Web Controller

I needed a lightweight web UI for easily switching between relay locations for a Mullvad client that's running on one of my servers. That server is proxying traffic for some of my devices. Occasionally, I need to quickly switch the current relay location, but SSH-ing to that server and making the switch with Mullvad CLI is a pain.

Note that I'm using this on **Debian** with **systemd**. If your setup is different, some changes might be needed.

> [!WARNING]
> **Do not expose this outside of your local network.** There is no authentication. Cross-origin requests that change the relay are rejected, so other websites can't switch it from your browser, but anyone who can reach the port can. In my case, I embed this UI in a Home Assistant dashboard.

<img width="831" src="https://github.com/user-attachments/assets/2e72ac28-52f4-4674-a97e-ded438a85ca7" />

## Installation

Having Mullvad with [the CLI](https://mullvad.net/en/help/how-use-mullvad-cli) installed is a prerequisite. You will
also need Go.

```shell
sudo apt install golang -y
```

Clone this repository. Build and install the systemd service:

```shell
make install
```

Running it again after pulling new changes rebuilds the binary and restarts the service.

### Removal

```shell
make uninstall
```

## Usage

```shell
systemctl status mullvad-web-controller
```

The URL to access the UI should be visible in the logs. By default it runs on port 8666.

## Configuration

Settings are read from environment variables:

| Variable              | Default        | Description                                                                 |
|-----------------------|----------------|-----------------------------------------------------------------------------|
| `MULLVAD_WEB_ADDR`    | `0.0.0.0:8666` | Address to listen on.                                                       |
| `MULLVAD_WEB_DRY_RUN` | `false`        | Set to `1` or `true` to use a fake Mullvad CLI. Useful for development.     |

The systemd service loads them from `/etc/default/mullvad-web-controller`, if that file exists. For example:

```shell
MULLVAD_WEB_ADDR=192.168.1.10:8666
```

Restart the service after changing it:

```shell
sudo systemctl restart mullvad-web-controller
```

## Working on it

Run a local copy in dry-run mode on http://localhost:8666:

```shell
make run-dev
```

Run the checks and tests:

```shell
make test
```
