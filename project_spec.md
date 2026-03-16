# UFactory module extension to expose by proxy UFactory Studio

UFactory robot arm controllers expose a web app for controlling the arm at <ipaddresss>:18333.

This work will extend the existing UFactory Viam module (https://github.com/viam-modules/viam-ufactory-xarm) to proxy that port on the Viam host where the module runs.

## Requirements
- Works on MacOS & linux. You can assume Viam-server runs as root.
- Packages any software that isn't standard on their respective hosts
- The proxy turns on if a "ufactory-studio-proxy" config field is set.
- also accepts a "ufactory-studio-proxy-port" optional field to set the local listening port (default 18333)

## Design Decisions
- The proxy config lives on the arm component (it already knows the controller IP)
- Proxy follows arm component lifecycle: starts on Reconfigure when enabled, stops on Close or when config field is removed
- No smart default port handling for multiple arms — both default to 18333 and whichever configures second gets an error
- No authentication on the proxy; security implications documented only
- Protocol support (HTTP, WebSocket, etc.) to be determined after testing on real hardware

## Documentation
- Update the docs
- Analyze and surface any security implications (no auth added, document the risk)


