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
//  1. PUT request to /api/caweb/(v1|v2)/settings
//     in order to update the mutable settings and reload the caweb.
//  2. GET request to /api/caweb/(v1|v2)/settings
//     in order to fetch the current visible settings.
//
// The v1 endpoints only deal with mutable settings themselves.
// The v2 endpoints also support the boundary values reporting.
func Register(r1, r2 *gin.RouterGroup, app *appuc.UseCase) {
	rs := &resource{app: app}
	r1.PUT("settings", rs.UpdateSettingsV1)
	r1.GET("settings", rs.FetchSettingsV1)
	r2.PUT("settings", rs.UpdateSettingsV2)
	r2.GET("settings", rs.FetchSettingsV2)
}

func (rs *resource) UpdateSettingsV1(c *gin.Context) {
	rs.UpdateSettings(c, false)
}

func (rs *resource) UpdateSettingsV2(c *gin.Context) {
	rs.UpdateSettings(c, true)
}

func (rs *resource) UpdateSettings(c *gin.Context, full bool) {
	req, ok := rs.DserUpdateSettingsReq(c)
	if !ok {
		return
	}
	vs, minb, maxb, err := rs.app.UpdateSettings(c, req)
	if err != nil {
		serdser.SerErr(c, err)
		return
	}
	if !full {
		c.JSON(http.StatusOK, vs)
		return
	}
	c.JSON(http.StatusOK, SettingsResp{
		Settings:  vs,
		MinBounds: minb,
		MaxBounds: maxb,
	})
}

func (rs *resource) FetchSettingsV1(c *gin.Context) {
	rs.FetchSettings(c, false)
}

func (rs *resource) FetchSettingsV2(c *gin.Context) {
	rs.FetchSettings(c, true)
}

func (rs *resource) FetchSettings(c *gin.Context, full bool) {
	vs, minb, maxb := rs.app.Settings()
	if !full {
		c.JSON(http.StatusOK, vs)
		return
	}
	c.JSON(http.StatusOK, SettingsResp{
		Settings:  vs,
		MinBounds: minb,
		MaxBounds: maxb,
	})
}
