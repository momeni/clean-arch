// Copyright (c) 2023 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package carsrs

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"github.com/momeni/clean-arch/pkg/adapter/restful/gin/serdser"
	"github.com/momeni/clean-arch/pkg/core/model"
)

type rawCarUpdateReq struct {
	Op   string         `form:"op" binding:"required,oneof=ride park"`
	Dst  *StrCoordinate `binding:"omitempty"`
	Mode string         `form:"mode" binding:"omitempty,oneof=old new"`
}

// StrCoordinate is a string-based representation (instead of a numeric
// representation) of a geographical location.
type StrCoordinate struct {
	Lat string `form:"lat" binding:"required,latitude"`
	Lon string `form:"lon" binding:"required,longitude"`
}

type carUpdateReq struct {
	CarID uuid.UUID
	Op    string
	Dst   model.Coordinate
	Mode  model.ParkingMode
}

// ToModel method converts a StrCoordinate to a model.Coordinate struct
// instance. Existence of this method allows all conversion codes to
// rely on exactly one implementation, so all fields should be listed
// here (and if some of them were missed after an update, fixing one
// place is enough to restore the whole project's sanity again).
func (sc StrCoordinate) ToModel() (c model.Coordinate, err error) {
	c.Lat, err = strconv.ParseFloat(sc.Lat, 64)
	if err != nil {
		return
	}
	c.Lon, err = strconv.ParseFloat(sc.Lon, 64)
	return
}

func (rs *resource) DserUpdateCarReq(c *gin.Context) *carUpdateReq {
	req := &rawCarUpdateReq{}
	val := &carUpdateReq{}
	if ok := serdser.Bind(c, req, binding.Form); !ok {
		return nil
	}
	var errs map[string][]string
	defer func() {
		if errs != nil {
			c.JSON(http.StatusBadRequest, errs)
		}
	}()
	var err error
	val.CarID, err = uuid.Parse(c.Param("cid"))
	if err != nil {
		serdser.AddErr(&errs, "cid", "Path param cid is not UUID.")
		return nil
	}
	val.Op = req.Op
	switch req.Op {
	case "ride":
		if serdser.Assert(
			&errs, req.Dst != nil,
			"lat/lon", "The op=ride requires lat and lon.",
		) && serdser.Assert(
			&errs, req.Mode == "",
			"mode", "The op=ride does not need mode.",
		) {
			val.Dst, err = req.Dst.ToModel()
			if err != nil {
				serdser.AddErr(&errs, "lat/lon", err.Error())
			}
		}
	case "park":
		if serdser.Assert(
			&errs, req.Dst == nil,
			"lat/lon", "The op=park does not need lat/lon.",
		) && serdser.Assert(
			&errs, req.Mode != "",
			"mode", "The op=park requires mode.",
		) {
			val.Mode, err = model.ParseParkingMode(req.Mode)
			if err != nil {
				serdser.AddErr(&errs, "mode", err.Error())
			}
		}
	default:
		panic("unknown op")
	}
	if errs == nil {
		return val
	}
	return nil
}
