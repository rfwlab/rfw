# Devtools

## Development Server

Run the development server:

```bash
rfw dev
```

### Flags

- `--port` specify port
- `--host` expose to network
- `--debug` enable verbose logs and profiling endpoints (`/debug/vars`, `/debug/pprof/`)

## Hot Reload

Changes to Go, RTML, Markdown and plugin assets trigger automatic rebuilds.
New directories are watched automatically.

## Profiling

With `--debug` flag, runtime metrics are available:

- `/debug/vars` exposes counters like `rebuilds`
- `/debug/pprof/` provides pprof profiles

