// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//nolint:paralleltest
//revive:disable:max-public-structs,unused-receiver,unused-parameter,function-length
package agent

import (
	"net"
	"testing"
)

func Test_address_Name(t *testing.T) {
	type fields struct {
		addr net.IP
	}
	tests := []struct {
		name   string
		want   string
		fields fields
	}{
		{
			name:   "ipv4",
			fields: fields{addr: net.ParseIP("192.168.1.1")},
			want:   "External IPv4 Address",
		},

		{
			name:   "ipv6",
			fields: fields{addr: net.ParseIP("2001:0db8:85a3:0000:0000:8a2e:0370:7334")},
			want:   "External IPv6 Address",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &address{
				addr: tt.fields.addr,
			}
			if got := a.Name(); got != tt.want {
				t.Errorf("address.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_address_ID(t *testing.T) {
	type fields struct {
		addr net.IP
	}
	tests := []struct {
		name   string
		want   string
		fields fields
	}{
		{
			name:   "ipv4",
			fields: fields{addr: net.ParseIP("192.168.1.1")},
			want:   "external_ipv4_address",
		},

		{
			name:   "ipv6",
			fields: fields{addr: net.ParseIP("2001:0db8:85a3:0000:0000:8a2e:0370:7334")},
			want:   "external_ipv6_address",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &address{
				addr: tt.fields.addr,
			}
			if got := a.ID(); got != tt.want {
				t.Errorf("address.ID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_address_Icon(t *testing.T) {
	type fields struct {
		addr net.IP
	}
	tests := []struct {
		name   string
		want   string
		fields fields
	}{
		{
			name:   "ipv4",
			fields: fields{addr: net.ParseIP("192.168.1.1")},
			want:   "mdi:numeric-4-box-outline",
		},

		{
			name:   "ipv6",
			fields: fields{addr: net.ParseIP("2001:0db8:85a3:0000:0000:8a2e:0370:7334")},
			want:   "mdi:numeric-6-box-outline",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &address{
				addr: tt.fields.addr,
			}
			if got := a.Icon(); got != tt.want {
				t.Errorf("address.Icon() = %v, want %v", got, tt.want)
			}
		})
	}
}
