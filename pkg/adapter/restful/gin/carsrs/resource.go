// Copyright (c) 2023-2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package carsrs realizes the cars resource, allowing the cars
// manipulation REST APIs to be accepted and delegated to the
// cars use cases respectively.
package carsrs

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/momeni/clean-arch/pkg/adapter/restful/gin/serdser"
	"github.com/momeni/clean-arch/pkg/core/model"
	"github.com/momeni/clean-arch/pkg/core/usecase/carsuc"
)

type resource struct {
	cars func() *carsuc.UseCase
}

// Register instantiates a resource adapting the cars use case instance
// with the relevant REST APIs including:
//  1. PATCH request to /api/caweb/(v1|v2)/cars/:cid
//     in order to ride or park a car.
//
// The same APIs are published as v1 and v2 RESTful endpoits.
func Register(r1, r2 *gin.RouterGroup, cars func() *carsuc.UseCase) {
	rs := &resource{cars: cars}
	r1.PATCH("cars/:cid", rs.UpdateCar)
	r2.PATCH("cars/:cid", rs.UpdateCar)
}

func (rs *resource) UpdateCar(c *gin.Context) {
	req, ok := rs.DserUpdateCarReq(c)
	if !ok {
		return
	}
	carsUseCase := rs.cars()
	var car *model.Car
	var err error
	switch req.Op {
	case "ride":
		car, err = carsUseCase.Ride(c, req.CarID, req.Dst)
	case "park":
		car, err = carsUseCase.Park(c, req.CarID, req.Mode)
	default:
		panic("unexpected op:" + req.Op)
	}
	if err != nil {
		serdser.SerErr(c, err)
		return
	}
	c.JSON(http.StatusOK, car)
}
