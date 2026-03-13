# UFactory module extension to expose by proxy UFactory Studio

UFactory robot arm controllers expose a web app for controlling the arm at <ipaddresss>:18333.

This work will extend the existing UFactory Viam module (https://github.com/viam-modules/viam-ufactory-xarm) to proxy that port on the Viam host where the module runs.

## Requirements
- Works on MacOS & linux. You can assume Viam-server runs as root.
- Packages any software that isn't standard on their respective hosts
- The proxy turns on if a "ufactory-studio-proxy" config field is set.
- also accepts a "ufactory-studio-proxy-port" optional field to set a port other than 18333

## Documentation
- Update the docs
- Analyze and surface any security implications


