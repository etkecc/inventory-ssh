# inventory-ssh

inventory-ssh is a lightweight wrapper around the standard `ssh` client. It reads `ansible.cfg` and Ansible inventory files from the current directory, resolves the requested host, and then runs SSH with the resolved options. If a host is not found, it can fall back to normal SSH behavior.

## Features

* Works as a drop-in wrapper for `ssh`
* Reads `ansible.cfg` and inventory in the current directory
* Supports defaults for user, port, and keys
* Optional "inventory only" mode to disable fallback
* Debug logging toggle

## Requirements

* A working `ssh` client (OpenSSH recommended)
* Ansible inventory files and `ansible.cfg` in the directory where you run the tool

## Install

### Binaries and distro packages

See the Releases page on GitHub for prebuilt binaries and packages.

### Build from source

```bash
just build
# or
go build .
```

## Configuration

Copy the sample config to your XDG config directory and rename it:

```bash
cp config.yml.sample "${XDG_CONFIG_HOME:-$HOME/.config}/inventory-ssh.yml"
```

Key options (see `config.yml.sample` for the full list):

* `path`: inventory file path (default `./hosts`)
* `ssh_command`: path or name of the ssh binary
* `inventory_only`: if `true`, do not fall back to plain SSH
* `debug`: enable debug logs
* `defaults`: default `user`, `port`, `private_keys`, and passwords

## Usage

Run it exactly like `ssh`, but from a directory containing your inventory:

```bash
inventory-ssh my-host
```

Optional: set an alias to transparently use it as `ssh`:

```bash
# $HOME/.bashrc
alias ssh="inventory-ssh"
```

## Examples

Connect using inventory resolution:

```bash
inventory-ssh app-server-1
```

Force inventory-only behavior (set in config):

```yaml
inventory_only: true
```

## Notes

* If the host is not found and `inventory_only` is `false`, inventory-ssh runs the underlying `ssh` command directly.
* Ansible is a trademark of Red Hat, Inc. This project is not affiliated with, endorsed by, or sponsored by Red Hat or the Ansible project.

## License

See `LICENSE.md`.
