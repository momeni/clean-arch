// Copyright (c) 2023 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package gin wraps the gin-gonic web framework and lists the
// middlewares which are expected to be enabled/disabled by the
// configuration settings.
package gin

import "github.com/gin-gonic/gin"

// HandlerFunc defines a gin middleware function.
// Each middleware will be called in order of its registration,
// may process the given gin.Context argument, pass control to
// the subsequent handlers by calling the ctx.Next() method,
// and then run some finalizing codes (if any).
type HandlerFunc = gin.HandlerFunc

// Engine represents the main object type containing all gin framework
// details such a middlewares and configurations.
type Engine = gin.Engine

// New creates a new gin Engine instance, registering the given
// middleware functions (if any).
func New(middlewares ...HandlerFunc) *Engine {
	e := gin.New()
	e.Use(middlewares...)
	return e
}

// Logger middleware logs incoming requests and their responses
// which is useful for debugging.
func Logger() HandlerFunc {
	return gin.Logger()
}

// Recovery middleware recovers panics and responds with 500 to clients.
func Recovery() HandlerFunc {
	return gin.Recovery()
}
