package carsrs

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/momeni/clean-arch/pkg/adapter/restful/gin/serdser"
	"github.com/momeni/clean-arch/pkg/core/model"
)

type rawCarUpdateReq struct {
	CarID string        `uri:"cid" binding:"required,uuid4"`
	Op    string        `form:"op" binding:"required,oneof=ride park"`
	Dst   StrCoordinate `binding:"omitempty"`
	Mode  string        `form:"mode" binding:"omitempty,oneof=old new"`
}

type StrCoordinate struct {
	Lat string `form:"lat" binding:"required,latitude`
	Lon string `form:"lon" binding:"required,longitude"`
}

type carUpdateReq struct {
	CarID uuid.UUID
	Op    string
	Dst   model.Coordinate
	Mode  model.ParkingMode
}

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
	if ok := serdser.Bind(c, req); !ok {
		return nil
	}
	var errs map[string][]string
	defer func() {
		if errs != nil {
			c.JSON(http.StatusBadRequest, errs)
		}
	}()
	var err error
	val.CarID, err = uuid.Parse(req.CarID)
	if err != nil {
		serdser.AddErr(&errs, "cid", "Path param cid is not UUID.")
		return nil
	}
	switch req.Op {
	case "ride":
		if serdser.Assert(&errs, req.Dst.Lat != "", "lat/lon", "The op=ride requires lat and lon.") &&
			serdser.Assert(&errs, req.Mode == "", "mode", "The op=ride does not need mode.") {
			val.Dst, err = req.Dst.ToModel()
			serdser.Assert(&errs, err == nil, "lat/lon", err.Error())
		}
	case "park":
		if serdser.Assert(&errs, req.Dst.Lat == "", "lat/lon", "The op=park does not need lat/lon.") &&
			serdser.Assert(&errs, req.Mode != "", "mode", "The op=park requires mode.") {
			val.Mode, err = model.ParseParkingMode(req.Mode)
			serdser.Assert(&errs, err == nil, "mode", err.Error())
		}
	default:
		panic("unknown op")
	}
	if errs == nil {
		return val
	}
	return nil
}
