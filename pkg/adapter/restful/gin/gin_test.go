// Copyright (c) 2023-2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package gin_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bitcomplete/sqltestutil"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/momeni/clean-arch/internal/test/dbcontainer"
	"github.com/momeni/clean-arch/pkg/adapter/config"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres"
	"github.com/momeni/clean-arch/pkg/adapter/restful/gin"
	"github.com/momeni/clean-arch/pkg/adapter/restful/gin/routes"
	"github.com/momeni/clean-arch/pkg/core/model"
	"github.com/momeni/clean-arch/pkg/core/repo"
	"github.com/stretchr/testify/suite"
)

type IntegrationGinTestSuite struct {
	suite.Suite

	Ctx  context.Context
	Pg   *sqltestutil.PostgresContainer
	Pool *postgres.Pool
	Gin  *gin.Engine
}

func TestIntegrationGinTestSuite(t *testing.T) {
	ctx := context.Background()
	pg, pool, dfrs, ok := dbcontainer.New(ctx, 60*time.Second, t)
	for _, f := range dfrs {
		defer f()
	}
	if !ok {
		return // errors are already logged
	}
	suite.Run(t, &IntegrationGinTestSuite{
		Ctx:  ctx,
		Pg:   pg,
		Pool: pool,
	})
}

func (igts *IntegrationGinTestSuite) SetupSuite() {
	sql, err := os.ReadFile("testdata/schema.sql")
	igts.Require().NoError(err, "failed to read schema.sql file")
	err = igts.Pool.Conn(
		igts.Ctx, func(ctx context.Context, c repo.Conn) error {
			_, err := c.Exec(ctx, string(sql))
			return err
		},
	)
	igts.Require().NoError(err, "failed to create schema contents")

	igts.Gin = gin.New(gin.Logger(), gin.Recovery())
	igts.Require().NotNil(igts.Gin, "cannot instantiate Gin engine")
	delay := 2 * time.Second
	err = routes.Register(igts.Gin, igts.Pool, config.Usecases{
		Cars: config.Cars{
			OldParkingDelay: &delay,
		},
	})
	igts.Require().NoError(err, "failed to register Gin routes")
}

func stringAddr(s string) *string {
	return &s
}

func urlEncoded(m map[string]string) io.Reader {
	u := url.Values{}
	for k, v := range m {
		u.Set(k, v)
	}
	return strings.NewReader(u.Encode())
}

func (igts *IntegrationGinTestSuite) TestBadRequest() {
	missingCarID := uuid.New()
	for _, tc := range []struct {
		name       string
		body       io.Reader
		detail, op *string
		lat, lon   *string
		latLon     *string `json:"lat/lon"`
		mode       *string
	}{
		{
			name:   "no body",
			body:   nil,
			detail: stringAddr("missing form body"),
		},
		{
			name: "empty body",
			body: urlEncoded(nil),
			op:   stringAddr("failed on the 'required' tag"),
		},
		{
			name: "invalid op",
			body: urlEncoded(map[string]string{
				"op": "invalid",
			}),
			op: stringAddr("failed on the 'oneof' tag"),
		},
		{
			name: "ride no-dst",
			body: urlEncoded(map[string]string{
				"op": "ride",
			}),
			latLon: stringAddr("op=ride requires lat and lon"),
		},
		{
			name: "ride no-lat",
			body: urlEncoded(map[string]string{
				"op":  "ride",
				"lon": "23",
			}),
			lat: stringAddr("failed on the 'required' tag"),
		},
		{
			name: "ride no-lon",
			body: urlEncoded(map[string]string{
				"op":  "ride",
				"lat": "40",
			}),
			lon: stringAddr("failed on the 'required' tag"),
		},
		{
			name: "ride with dst and mode",
			body: urlEncoded(map[string]string{
				"op":   "ride",
				"lon":  "23",
				"lat":  "40",
				"mode": "new",
			}),
			mode: stringAddr("op=ride does not need mode"),
		},
		{
			name: "park no-mode",
			body: urlEncoded(map[string]string{
				"op": "park",
			}),
			mode: stringAddr("op=park requires mode"),
		},
		{
			name: "park with dst and mode",
			body: urlEncoded(map[string]string{
				"op":   "park",
				"lon":  "23",
				"lat":  "40",
				"mode": "new",
			}),
			latLon: stringAddr("op=park does not need lat/lon"),
		},
		{
			name: "park with invalid mode",
			body: urlEncoded(map[string]string{
				"op":   "park",
				"mode": "invalid",
			}),
			mode: stringAddr("failed on the 'oneof' tag"),
		},
	} {
		igts.Run(tc.name, func() {
			w := httptest.NewRecorder()
			req, err := http.NewRequest(
				http.MethodPatch,
				"/api/caweb/v1/cars/"+missingCarID.String(),
				tc.body,
			)
			igts.Require().NoError(err, "cannot create PATCH request")

			res := &struct {
				Detail   string
				Op       []string
				Lat, Lon []string
				LatLon   []string `json:"lat/lon"`
				Mode     []string
			}{}
			igts.sendReqRecvResp(w, req, res)

			igts.Equal(400, w.Code)
			if tc.detail != nil {
				igts.Equal(*tc.detail, res.Detail, "wrong detail")
			}
			igts.assertOptContains(tc.op, res.Op, "wrong op")
			igts.assertOptContains(tc.lat, res.Lat, "wrong lat")
			igts.assertOptContains(tc.lon, res.Lon, "wrong lon")
			igts.assertOptContains(tc.latLon, res.LatLon, "wrong lat/lon")
			igts.assertOptContains(tc.mode, res.Mode, "wrong mode")
		})
	}
}

func (igts *IntegrationGinTestSuite) sendReqRecvResp(
	w *httptest.ResponseRecorder, req *http.Request, res any,
) {
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	igts.Gin.ServeHTTP(w, req)
	b := w.Body.Bytes()
	igts.NoError(json.Unmarshal(b, res), "body is not json")
}

func (igts *IntegrationGinTestSuite) assertOptContains(
	expectedPart *string, seen []string, msgAndArgs ...any,
) bool {
	if expectedPart == nil {
		return true
	}
	if !igts.Equal(1, len(seen), msgAndArgs...) {
		return false
	}
	return igts.Contains(seen[0], *expectedPart, msgAndArgs...)
}

func (igts *IntegrationGinTestSuite) TestNotFound() {
	missingCarID := uuid.New()
	for _, tc := range []struct {
		name string
		body io.Reader
	}{
		{
			name: "ride",
			body: urlEncoded(map[string]string{
				"op":  "ride",
				"lon": "23",
				"lat": "40",
			}),
		},
		{
			name: "park",
			body: urlEncoded(map[string]string{
				"op":   "park",
				"mode": "new",
			}),
		},
	} {
		igts.Run(tc.name, func() {
			w := httptest.NewRecorder()
			req, err := http.NewRequest(
				http.MethodPatch,
				"/api/caweb/v1/cars/"+missingCarID.String(),
				tc.body,
			)
			igts.Require().NoError(err, "cannot create PATCH request")

			res := &struct {
				Detail string
			}{}
			igts.sendReqRecvResp(w, req, res)

			igts.Equal(404, w.Code)
			igts.Equal(
				"expected one row, but got 0", res.Detail,
				"wrong detail",
			)
		})
	}
}

func (igts *IntegrationGinTestSuite) createCar(car *model.Car) (
	uuid.UUID, error,
) {
	carID := uuid.New()
	err := igts.Pool.Conn(
		igts.Ctx, func(ctx context.Context, c repo.Conn) error {
			count, err := c.Exec(
				ctx,
				`INSERT INTO cars(cid, name, lat, lon, parked)
VALUES ($1, $2, $3, $4, $5)`,
				carID, car.Name,
				car.Coordinate.Lat, car.Coordinate.Lon,
				car.Parked,
			)
			igts.Equal(int64(1), count, "tried to INSERT one car")
			return err
		},
	)
	return carID, err
}

func (igts *IntegrationGinTestSuite) TestRide() {
	carID, err := igts.createCar(&model.Car{
		Name: "test-car",
		Coordinate: model.Coordinate{
			Lat: 10.1,
			Lon: 12.2,
		},
		Parked: true,
	})
	igts.Require().NoError(err, "failed to create initial car in DB")
	w := httptest.NewRecorder()
	req, err := http.NewRequest(
		http.MethodPatch,
		"/api/caweb/v1/cars/"+carID.String(),
		urlEncoded(map[string]string{
			"op":  "ride",
			"lon": "10.5",
			"lat": "15.9",
		}),
	)
	igts.Require().NoError(err, "cannot create PATCH request")

	res := &model.Car{}
	igts.sendReqRecvResp(w, req, res)

	igts.Equal(200, w.Code)
	igts.Equal(
		model.Car{
			Name: "test-car",
			Coordinate: model.Coordinate{
				Lat: 15.9,
				Lon: 10.5,
			},
			Parked: false,
		},
		*res,
		"unexpected resulting car instance",
	)
}

func (igts *IntegrationGinTestSuite) TestPark() {
	carID, err := igts.createCar(&model.Car{
		Name: "test-car",
		Coordinate: model.Coordinate{
			Lat: 10.2,
			Lon: 12.3,
		},
		Parked: false,
	})
	igts.Require().NoError(err, "failed to create initial car in DB")
	w := httptest.NewRecorder()
	req, err := http.NewRequest(
		http.MethodPatch,
		"/api/caweb/v1/cars/"+carID.String(),
		urlEncoded(map[string]string{
			"op":   "park",
			"mode": "new",
		}),
	)
	igts.Require().NoError(err, "cannot create PATCH request")

	res := &model.Car{}
	igts.sendReqRecvResp(w, req, res)

	igts.Equal(200, w.Code)
	igts.Equal(
		model.Car{
			Name: "test-car",
			Coordinate: model.Coordinate{
				Lat: 10.2,
				Lon: 12.3,
			},
			Parked: true,
		},
		*res,
		"unexpected resulting car instance",
	)
}
