# Env Plugin

The **Env plugin** makes environment variables available to your rfw app at build time. It collects variables prefixed with `RFW_` and generates a temporary Go package (`rfwenv`) that exposes them through a simple getter.

## Features

* Reads all environment variables starting with `RFW_`.
* Generates a temporary `rfwenv` package with a `Get(key)` function.
* Cleans up the generated code after the build finishes.
* Keeps values available in Go code without manual parsing.

## Usage

1. Set environment variables with the prefix `RFW_`.

   ```sh
   export RFW_API_URL=https://api.example.com
   export RFW_FEATURE_FLAG=true
   ```

2. Register the plugin in your app (usually automatic):

   ```go
   import (
       core "github.com/rfwlab/rfw/v1/core"
       env "github.com/rfwlab/rfw/v1/plugins/env"
   )

   func main() {
       core.RegisterPlugin(&env.plugin{})
   }
   ```

3. Use the generated package in your Go code:

   ```go
   import "rfwenv"

   func main() {
       api := rfwenv.Get("API_URL")
       println("API URL:", api)
   }
   ```

## API Reference

The generated package contains:

| Function                        | Description                                                        |
| ------------------------------- | ------------------------------------------------------------------ |
| `rfwenv.Get(key string) string` | Returns the value of the given key, or an empty string if not set. |

## Notes

* Only variables starting with `RFW_` are included. The prefix is stripped in `rfwenv.Get`.
* The generated package exists only during the build and is deleted afterward.
* If two variables share the same key (case-sensitive), the last one defined in the environment will be used.
* Useful for injecting build-time configuration such as API endpoints, flags, or keys.
