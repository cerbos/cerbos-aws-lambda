// Copyright 2021 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"testing"
	"time"
)

func Test_parseDurationOrDefault(t *testing.T) {
	defaultValue := time.Duration(time.Now().Second()) * time.Nanosecond
	tests := []struct {
		value string
		want  time.Duration
	}{
		// valid values
		{value: "10s", want: 10 * time.Second},
		{value: "31ms", want: 31 * time.Millisecond},
		// if invalid values return default
		{value: "", want: defaultValue},
		{value: "hi", want: defaultValue},
		{value: "-1", want: defaultValue},
		{value: "0x123", want: defaultValue},
		{value: "4ss", want: defaultValue},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			if got := parseDurationOrDefault(tt.value, defaultValue); got != tt.want {
				t.Errorf("parseDurationOrDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}
