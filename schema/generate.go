// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

//revive:disable:package-comments
package api

//go:generate go tool oapi-codegen -config models-cfg.yaml models.yaml
//go:generate go tool oapi-codegen -config rest-cfg.yaml rest.yaml
