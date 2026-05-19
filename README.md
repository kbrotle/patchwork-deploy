# patchwork-deploy

Minimal deployment orchestrator for small VPS setups using SSH and declarative config files.

---

## Installation

```bash
go install github.com/yourusername/patchwork-deploy@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/patchwork-deploy.git
cd patchwork-deploy && go build -o patchwork-deploy .
```

---

## Usage

Define your deployment in a `deploy.yml` file:

```yaml
hosts:
  - address: user@192.168.1.10
    key: ~/.ssh/id_ed25519

steps:
  - name: pull latest
    run: git -C /srv/myapp pull origin main

  - name: restart service
    run: systemctl restart myapp
```

Then run:

```bash
patchwork-deploy run --config deploy.yml
```

To do a dry run without executing remote commands:

```bash
patchwork-deploy run --config deploy.yml --dry-run
```

---

## Features

- Declarative YAML config
- SSH key authentication
- Sequential step execution with early failure
- Dry-run mode
- No agents, no daemons — just SSH

---

## Requirements

- Go 1.21+
- SSH access to target hosts

---

## License

MIT © 2024 yourusername