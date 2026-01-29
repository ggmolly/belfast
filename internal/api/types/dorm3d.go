package types

import "github.com/ggmolly/belfast/internal/orm"

type Dorm3dApartment = orm.Dorm3dApartment

type Dorm3dApartmentRequest = orm.Dorm3dApartment

type Dorm3dGiftList = orm.Dorm3dGiftList

type Dorm3dShipList = orm.Dorm3dShipList

type Dorm3dRoomList = orm.Dorm3dRoomList

type Dorm3dInsList = orm.Dorm3dInsList

type Dorm3dApartmentListResponse struct {
	Apartments []Dorm3dApartment `json:"apartments"`
	Meta       PaginationMeta    `json:"meta"`
}
