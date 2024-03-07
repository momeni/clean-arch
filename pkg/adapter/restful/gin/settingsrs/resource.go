// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package settingsrs realizes the settings resource, allowing the
// settings fetching and replacement (mutation) REST APIs to be accepted
// and delegated to the application use case properly.
package settingsrs

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/momeni/clean-arch/pkg/adapter/restful/gin/serdser"
	"github.com/momeni/clean-arch/pkg/core/usecase/appuc"
)

type resource struct {
	app *appuc.UseCase
}

// Register instantiates a resource adapting the app use case instance
// with the relevant REST APIs including:
//  1. PUT request to /api/caweb/v1/settings
//     in order to update the mutable settings and reload the caweb.
//  2. GET request to /api/caweb/v1/settings
//     in order to fetch the current visible settings.
func Register(r *gin.RouterGroup, app *appuc.UseCase) {
	rs := &resource{app: app}
	r.PUT("settings", rs.UpdateSettings)
	r.GET("settings", rs.FetchSettings)
}

func (rs *resource) UpdateSettings(c *gin.Context) {
	req, ok := rs.DserUpdateSettingsReq(c)
	if !ok {
		return
	}
	vs, err := rs.app.UpdateSettings(c, req)
	if err != nil {
		serdser.SerErr(c, err)
		return
	}
	c.JSON(http.StatusOK, vs)
}

func (rs *resource) FetchSettings(c *gin.Context) {
	vs := rs.app.Settings()
	c.JSON(http.StatusOK, vs)
}
