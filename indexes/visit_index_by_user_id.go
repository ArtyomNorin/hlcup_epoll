package indexes

import (
	"hlcup_epoll/entities"
	"sync"
		)

type VisitIndexByUserId struct {
	visits map[uint][]uint
	mutex  *sync.Mutex
}

func NewVisitIndexByUserId() *VisitIndexByUserId {
	return &VisitIndexByUserId{visits: make(map[uint][]uint), mutex: new(sync.Mutex)}
}

func (visitIndexByUserId *VisitIndexByUserId) AddVisit(visit *entities.Visit) {

	visitIndexByUserId.mutex.Lock()

	visitIndexByUserId.visits[*visit.User] = append(visitIndexByUserId.visits[*visit.User], *visit.Id)

	visitIndexByUserId.mutex.Unlock()
}

func (visitIndexByUserId *VisitIndexByUserId) GetVisits(userId uint) []uint {

	visits, isIdExist := visitIndexByUserId.visits[userId]

	if !isIdExist {
		return nil
	}

	return visits
}

func (visitIndexByUserId *VisitIndexByUserId) DeleteVisit(userId uint, visitId uint) {

	visitIndexByUserId.mutex.Lock()

	visitsByUserId, isUserExist := visitIndexByUserId.visits[userId]

	if !isUserExist || len(visitsByUserId) == 0 {
		visitIndexByUserId.mutex.Unlock()
		return
	}

	for visitIndex, visitValue := range visitsByUserId {

		if visitValue == visitId {
			visitIndexByUserId.visits[userId] = append(visitsByUserId[:visitIndex], visitsByUserId[visitIndex+1:]...)
			break
		}
	}

	visitIndexByUserId.mutex.Unlock()
}
