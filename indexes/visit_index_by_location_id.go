package indexes

import (
	"hlcup_epoll/entities"
	"sync"
	)

type VisitIndexByLocationId struct {
	visits map[uint][]uint
	mutex  *sync.Mutex
}

func NewVisitIndexByLocationId() *VisitIndexByLocationId {
	return &VisitIndexByLocationId{visits: make(map[uint][]uint), mutex: new(sync.Mutex)}
}

func (visitIndexByLocationId *VisitIndexByLocationId) AddVisit(visit *entities.Visit) {

	visitIndexByLocationId.mutex.Lock()

	visitIndexByLocationId.visits[*visit.Location] = append(visitIndexByLocationId.visits[*visit.Location], *visit.Id)

	visitIndexByLocationId.mutex.Unlock()
}

func (visitIndexByLocationId *VisitIndexByLocationId) GetVisits(locationId uint) []uint {

	visits, isIdExist := visitIndexByLocationId.visits[locationId]

	if !isIdExist {
		return nil
	}

	return visits
}

func (visitIndexByLocationId *VisitIndexByLocationId) DeleteVisit(locationId uint, visitId uint) {

	visitIndexByLocationId.mutex.Lock()

	visitsByLocationId, isLocationExist := visitIndexByLocationId.visits[locationId]

	if !isLocationExist || len(visitsByLocationId) == 0 {
		visitIndexByLocationId.mutex.Unlock()
		return
	}

	for visitIndex, visitValue := range visitsByLocationId {

		if visitValue == visitId {
			visitIndexByLocationId.visits[locationId] = append(visitsByLocationId[:visitIndex], visitsByLocationId[visitIndex+1:]...)
			break
		}
	}

	visitIndexByLocationId.mutex.Unlock()
}
