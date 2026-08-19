# Lyra Board

| Item | Value |
|---|---:|
| Dashboard port | 6010 |
| SSH port | 6110 |

Public dashboard endpoint:

```text
http://34.87.151.202:6010/
```

Runtime transport:

- REST API: `/api/*`
- SSE stream: `/events`

Build:

```bash
./scripts/build_lyra.sh all
```

Artifacts:

```text
dist/lyra/go-backend
dist/lyra/dashboard-web/
dist/lyra/build-info.txt
```
