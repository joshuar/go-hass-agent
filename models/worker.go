// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package models

func (w *WorkerMetadata) ID() string {
	return w.WorkerID
}

func (w *WorkerMetadata) Description() string {
	return w.WorkerDescription
}

func SetWorkerMetadata(id, desc string) *WorkerMetadata {
	return &WorkerMetadata{WorkerID: id, WorkerDescription: desc}
}
