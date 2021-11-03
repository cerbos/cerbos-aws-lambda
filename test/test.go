// Copyright 2021 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"path/filepath"
	"runtime"
	"testing"
)

func PathToDir(tb testing.TB, dir string) string {
	tb.Helper()

	_, currFile, _, ok := runtime.Caller(0)
	if !ok {
		tb.Error("Failed to detect testdata directory")
		return ""
	}

	return filepath.Join(filepath.Dir(currFile), "testdata", dir)
}
