package carsrs

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/momeni/clean-arch/pkg/adapter/restful/gin/serdser"
	"github.com/momeni/clean-arch/pkg/core/model"
	"github.com/momeni/clean-arch/pkg/core/usecase/carsuc"
)

type resource struct {
	cars *carsuc.UseCase
}

func Register(r *gin.RouterGroup, cars *carsuc.UseCase) {
	rs := &resource{cars: cars}
	r.PATCH("cars/:cid", rs.UpdateCar)
}

func (rs *resource) UpdateCar(c *gin.Context) {
	req := rs.DserUpdateCarReq(c)
	if req == nil {
		return
	}
	var car *model.Car
	var err error
	switch req.Op {
	case "ride":
		car, err = rs.cars.Ride(c, req.CarID, req.Dst)
	case "park":
		car, err = rs.cars.Park(c, req.CarID, req.Mode)
	default:
		panic("unexpected op:" + req.Op)
	}
	if err != nil {
		serdser.SerErr(c, err)
		return
	}
	c.JSON(http.StatusOK, car)
}
