package model

// Coordinate represents a geographical location with a latitude and
// longitude. This struct is included in the Car struct in order to
// demonstrate how a struct may be embedded while mapping them to a
// database table.
type Coordinate struct {
	Lat, Lon float64 // latitude and longitude of the geo-location
}
