package inventory

import "errors"

// Sentinel errors for the inventory subsystem.
// All error values are pre-allocated to prevent hot-path allocation.
var (
	ErrInventoryExhausted = errors.New("inventory exhausted")
	ErrInventoryMissing   = errors.New("inventory key missing")
	ErrLuaExecution       = errors.New("lua execution failure")
	ErrQueueFull          = errors.New("worker queue saturated")
)
