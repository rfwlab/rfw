# plugin

Interfaces for extending the framework during build and runtime.

| Method | Description |
| --- | --- |
| `PreBuild(cfg json.RawMessage) error` | Invoked before compilation begins. |
| `Build(cfg json.RawMessage) error` | Runs during the build step. |
| `PostBuild(cfg json.RawMessage) error` | Runs after the build completes. |
| `Install(a *core.App)` | Registers runtime features before the app starts. |
| `Uninstall(a *core.App)` | Cleans up resources added during `Install`. |
| `Name() string` | Optional identifier provided via the `Named` interface. |
| `ShouldRebuild(path string) bool` | Signals if a file change triggers a rebuild. |
| `Priority() int` | Execution order – lower numbers run first. |

