# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

Refer to @project_spec.md for details regarding any currently ongoing implementation.

## Build & Test Commands

```bash
make build              # Build binary to bin/viam-xarm
make module             # Build binary + package tarball (bin/module.tar.gz) with URDF/mesh assets
make test               # Run all tests with -race -failfast
make lint               # Format code (gofmt) + run golangci-lint with etc/.golangci.yaml
make tool-install       # Install golangci-lint and other tools to bin/gotools/
```

Run a single test:
```bash
go test -v -race -run TestFunctionName ./arm/
```

## Architecture

This is a **Viam module** (`viam:ufactory`) that integrates UFactory xArm robotic arms and grippers with the Viam platform. It registers 7 resource models in `main.go`:

- **4 arm models**: xArm6, xArm7, Lite6, xArm850 (all share the same `xArm` struct implementation)
- **3 gripper models**: standard gripper, lite gripper (Lite6 two-finger), vacuum gripper

All component logic lives in the `arm/` package. There is no sub-package hierarchy.

### Key Files

| File | Purpose |
|------|---------|
| `arm/xarm.go` | Arm component: config validation, Reconfigure, MoveToJointPositions, MoveThroughJointPositions, DoCommand, kinematics loading |
| `arm/comm.go` | Low-level TCP protocol (Modbus-like on port 502), register commands, error/state management, trajectory interpolation (`createRawJointSteps`) |
| `arm/gripper.go` | Standard and Lite gripper components (both share this file) |
| `arm/vacuum_gripper.go` | Vacuum gripper component |

### Communication Protocol

The arm communicates over TCP using a Modbus-like binary protocol. Frame format: `TID(2) | PROT(2) | LEN(2) | REG(1) | PARAMS(...)`. All protocol commands go through `comm.go`. The `moveLock` mutex serializes access to the TCP connection.

### Motion Modes

The arm supports three motion modes set via `setMotionMode()`:
- **Mode 0** (Position): Normal joint position control
- **Mode 1** (Servoj): Fast streaming for dense trajectory playback
- **Mode 2** (Manual): Gravity-compensated teaching mode

### Trajectory Planning

`createRawJointSteps()` in `comm.go` generates trapezoidal velocity profiles between waypoints at `moveHZ` frequency (default 100Hz). An optional external trajectory generator (ML model service like `trajex`) can replace the built-in interpolator.

### Kinematics Models

Four arm models each have embedded JSON kinematics files (compiled into the binary) and optional URDF files (loaded from disk). The `use_urdfs` config flag switches between them. 3D mesh assets in `arm/3d_models/` and `arm/meshes/` are packaged in the module tarball.

### DoCommand Extensions

The `DoCommand` method on both arms and grippers is the extensibility point for operations not covered by the standard Viam API: `set_speed`, `set_acceleration`, `load` (torques), `enter_manual_mode`, `exit_manual_mode`, `clear_error`, `get_state`, `get_error`, gripper position control, and vacuum control.

## Linting

Config is at `etc/.golangci.yaml`. Line length limit is 150 characters. The linter runs with `--fix` to auto-correct where possible.

## External References

- [xArm Developer Manual](https://www.ufactory.cc/wp-content/uploads/2023/04/xArm-Developer-Manual-V1.10.0.pdf) — protocol register map and command reference
- xArm Studio accessible at `http://<arm-ip>:18333` for diagnostics
