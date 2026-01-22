package types

import "github.com/ggmolly/belfast/internal/orm"

type Dorm3dApartment = orm.Dorm3dApartment

type Dorm3dApartmentRequest = orm.Dorm3dApartment

type Dorm3dApartmentListResponse struct {
	Apartments []Dorm3dApartment `json:"apartments"`
	Meta       PaginationMeta    `json:"meta"`
}
