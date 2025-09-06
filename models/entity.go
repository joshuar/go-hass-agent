// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package models

// Valid returns whether the entity contains valid data. This checks only
// whether the entity data is empty. To check validity of a specific type of
// entity, the data should extracted (with an As* method) and then the Valid
// method called on the data type.
func (e *Entity) Valid() bool {
	return e.union != nil
}
