// Copyright (c) 2026 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//go:generate go tool stringer -type=gpuVendorType -output gpuVendorType.gen.go -linecomment

package mem

// All of the possible GPU vendors, this is used to find the correct way to find these 
// memory statistics. We map these to an iota so we can use these values as settings
const (
	GPUVendorAMD          gpuVendorType = iota	// AMD 
	GPUVendorNvidia					// Nvidia 
	GPUVendorIntel					// Intel 
	GPUVendorNone					// None
)

type gpuVendorType int

// gpuVendorTypeNames maps the names of our settings back to the enum value.
var gpuVendorTypeNames = map[string]gpuVendorType{
	"AMD":           GPUVendorAMD,
	"Nvidia":        GPUVendorNvidia,
	"Intel":         GPUVendorIntel,
	"None":          GPUVendorNone,
}

