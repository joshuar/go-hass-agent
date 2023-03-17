# Home Assistant Test Server in Container via Podman

This directory contains a way to fire up a Home Assistant server running in a container via Podman.


### Configuration

- Set the Home Assistant version and your local timezone in `group_vars/all/01-general.yml`.

## Usage

```shell
ansible-playbook ./playbook.yml
```
