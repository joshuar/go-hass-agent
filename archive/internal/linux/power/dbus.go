// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package power

const (
	loginBasePath       = "/org/freedesktop/login1"
	loginSessionPath    = loginBasePath + "/Session"
	loginBaseInterface  = "org.freedesktop.login1"
	managerInterface    = loginBaseInterface + ".Manager"
	sessionInterface    = loginBaseInterface + ".Session"
	sessionLockSignal   = "Lock"
	sessionUnlockSignal = "Unlock"
	sessionLockedProp   = "LockedHint"
	sessionIdleProp     = "IdleHint"
	sessionIdleTimeProp = "IdleSinceHint"
)
