package indexes

import (
		"hlcup_epoll/entities"
	"sync"
	"github.com/json-iterator/go"
)

type VisitIndexById struct {
	visits map[uint][]byte
	mutex  *sync.Mutex
}

func NewVisitIndexById() *VisitIndexById {
	return &VisitIndexById{visits: make(map[uint][]byte), mutex: new(sync.Mutex)}
}

func (visitIndexById *VisitIndexById) AddVisit(visit *entities.Visit) error {

	encodedVisit, err := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(visit)

	if err != nil {
		return err
	}

	visitIndexById.mutex.Lock()

	visitIndexById.visits[*visit.Id] = encodedVisit

	visitIndexById.mutex.Unlock()

	return nil
}

func (visitIndexById *VisitIndexById) GetVisit(visitId uint) []byte {

	visitBytes, isIdExist := visitIndexById.visits[visitId]

	if !isIdExist {
		return nil
	}

	return visitBytes
}
