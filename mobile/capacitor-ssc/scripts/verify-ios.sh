#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${DEVELOPER_DIR:-}" && -d /Applications/Xcode.app/Contents/Developer ]]; then
	export DEVELOPER_DIR=/Applications/Xcode.app/Contents/Developer
fi

device_id="${RFW_IOS_SIMULATOR_ID:-}"
if [[ -z "$device_id" ]]; then
	device_id="$({
		xcrun simctl list devices available
	} | sed -nE 's/.*\(([0-9A-Fa-f-]{36})\) \((Booted|Shutdown)\).*/\1/p' | head -n 1)"
fi
if [[ -z "$device_id" ]]; then
	echo "No available iOS Simulator found; install one in Xcode or set RFW_IOS_SIMULATOR_ID." >&2
	exit 1
fi

exec xcodebuild \
	-scheme RfwlabCapacitorSsc \
	-destination "platform=iOS Simulator,id=$device_id" \
	test
