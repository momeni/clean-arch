// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package settingsrs

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/momeni/clean-arch/pkg/adapter/restful/gin/serdser"
	"github.com/momeni/clean-arch/pkg/core/model"
)

func (rs *resource) DserUpdateSettingsReq(
	c *gin.Context,
) (*model.Settings, bool) {
	req := &model.Settings{}
	if ok := serdser.Bind(c, req, binding.JSON); !ok {
		return nil, false
	}
	return req, true
}

// SettingsResp publishes three fields in order to be serialized as
// JSON fields and reported to the frontend as follows:
//  1. The settings field for reporting of visible settings which may
//     be mutable or immutable,
//  2. The min_bounds field for reporting the minimum acceptable value
//     for settings, all settings including the invisible items but
//     excluding those settings which do not have a known lower bound,
//  3. The max_bounds field for reporting the maximum acceptable value
//     for settings, all settings including the invisible items but
//     excluding those settings which do not have a known upper bound.
type SettingsResp struct {
	Settings  *model.VisibleSettings `json:"settings"`
	MinBounds *model.Settings        `json:"min_bounds"`
	MaxBounds *model.Settings        `json:"max_bounds"`
}
