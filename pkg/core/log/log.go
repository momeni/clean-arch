// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package log provides helper function over the standard log/slog
// structured logging package. By default, slog package-level functions
// such as slog.Info accept a message and a series of interleaved key
// and value arguments as a series of "any" arguments. A more efficient
// API is also provided which takes slog.Attr arguments which are typed
// statically and avoid memory allocation for simple data types.
// This log package exports Debug, Info, Warn, and Error functions
// which accept a context, message, and a series of slog.Attr arguments
// facilitating usage of the slog.LogAttrs function.
// It also provides helper functions for preparing slog.Attr instances.
package log

import (
	"context"
	"log/slog"
	"runtime"
	"time"
)

// Debug logs msg and attrs with the given context at the debug level.
func Debug(ctx context.Context, msg string, attrs ...slog.Attr) {
	logAttrs(ctx, slog.LevelDebug, msg, attrs...)
}

// Info logs msg and attrs with the given context at the info level.
func Info(ctx context.Context, msg string, attrs ...slog.Attr) {
	logAttrs(ctx, slog.LevelInfo, msg, attrs...)
}

// Warn logs msg and attrs with the given context at the warning level.
func Warn(ctx context.Context, msg string, attrs ...slog.Attr) {
	logAttrs(ctx, slog.LevelWarn, msg, attrs...)
}

// Error logs msg and attrs with the given context at the error level.
func Error(ctx context.Context, msg string, attrs ...slog.Attr) {
	logAttrs(ctx, slog.LevelError, msg, attrs...)
}

// logAttrs logs the msg and given attrs using the level log-level.
// It ignores the direct caller of logAttrs function when looking for
// its caller file name and line number, hence, it must be either
// exported and only called by client codes or non-exported and caller
// from this package itself. And since it is called from this package,
// it has to be non-exported.
func logAttrs(
	ctx context.Context,
	level slog.Level,
	msg string,
	attrs ...slog.Attr,
) {
	l := slog.Default()
	if !l.Enabled(ctx, level) {
		return
	}
	var pcs [1]uintptr
	// skip [runtime.Callers, this function, its parent in log pkg]
	runtime.Callers(3, pcs[:])
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	r.AddAttrs(attrs...)
	_ = l.Handler().Handle(ctx, r)
}
