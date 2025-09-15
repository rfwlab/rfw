# rtml

Helpers for exposing plugin data to templates.

| Item | Description |
| --- | --- |
| `RegisterRTMLVar(plugin, name string, val any)` | Make a value available in RTML as `{plugin:NAME.VAR}`. |
| `RegisterPluginVar(plugin, name string, val any)` | Convenience wrapper for plugins to expose RTML variables. |

