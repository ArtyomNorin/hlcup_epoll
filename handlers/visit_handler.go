package handlers

import (
	"hlcup_epoll/services"
	"log"
	"strconv"
	"github.com/json-iterator/go"
	"hlcup_epoll/entities"
	"net/http"
)

type VisitApiHandler struct {
	storage    *services.Storage
	errLogger  *log.Logger
	infoLogger *log.Logger
}

var oldLocationId, oldUserId uint

func NewVisitApiHandler(storage *services.Storage, errLogger *log.Logger, infoLogger *log.Logger) *VisitApiHandler {

	return &VisitApiHandler{storage: storage, errLogger: errLogger, infoLogger: infoLogger}
}

func (visitApiHandler *VisitApiHandler) GetById(request *http.Request, visitIdString string) ([]byte, int) {

	visitId, err := strconv.Atoi(visitIdString)

	if err != nil {
		return nil, 404
	}

	visit := visitApiHandler.storage.GetVisitById(uint(visitId))

	if visit == nil {
		return nil, 404
	} else {
		return visit, 200
	}
}

func (visitApiHandler *VisitApiHandler) Update(request *http.Request, visitIdString string) ([]byte, int) {

	visitId, err := strconv.Atoi(visitIdString)

	if err != nil {
		return nil, 404
	}

	visitBytes := visitApiHandler.storage.GetVisitById(uint(visitId))

	if visitBytes == nil {
		return nil, 404
	}

	newVisitMap := make(map[string]interface{})

	err = jsoniter.NewDecoder(request.Body).Decode(&newVisitMap)

	if err != nil {
		return nil, 400
	}

	visit := new(entities.Visit)

	err = jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(visitBytes, visit)

	if err != nil {
		visitApiHandler.errLogger.Fatalln(err)
	}

	if value, ok := newVisitMap["location"]; ok {

		locationId, typeOk := value.(float64)

		if value == nil || !typeOk || locationId <= 0 {
			return nil, 400
		}

		locationIdAsUint := uint(locationId)

		oldLocationId = *visit.Location
		visit.Location = &locationIdAsUint
	}

	if value, ok := newVisitMap["user"]; ok {

		userId, typeOk := value.(float64)

		if value == nil || !typeOk || userId <= 0 {
			return nil, 400
		}

		userIdAsUint := uint(userId)

		oldUserId = *visit.User
		visit.User = &userIdAsUint
	}

	if value, ok := newVisitMap["visited_at"]; ok {

		visitedAt, typeOk := value.(float64)

		if value == nil || !typeOk {
			return nil, 400
		}

		visitedAtAsInt := int(visitedAt)

		visit.VisitedAt = &visitedAtAsInt
	}

	if value, ok := newVisitMap["mark"]; ok {

		mark, typeOk := value.(float64)

		if value == nil || !typeOk || (mark != 0 && mark != 1 && mark != 2 && mark != 3 && mark != 4 && mark != 5) {
			return nil, 400
		}

		markAsInt := int(mark)

		visit.Mark = &markAsInt
	}

	if _, ok := newVisitMap["location"]; ok {
		visitApiHandler.storage.DeleteVisitFromLocation(oldLocationId, *visit.Id)
		visitApiHandler.storage.AddVisitByLocationId(visit)
	}

	if _, ok := newVisitMap["user"]; ok {
		visitApiHandler.storage.DeleteVisitFromUser(oldUserId, *visit.Id)
		visitApiHandler.storage.AddVisitByUserId(visit)
	}

	visitApiHandler.storage.AddVisit(visit)

	return []byte("{}"), 200
}

func (visitApiHandler *VisitApiHandler) Create(request *http.Request) ([]byte, int) {

	newVisitMap := make(map[string]interface{})

	err := jsoniter.NewDecoder(request.Body).Decode(&newVisitMap)

	if err != nil {
		return nil, 400
	}

	visitIdInterface, ok := newVisitMap["id"]

	if !ok {
		return nil, 400
	}

	visitIdFloat, typeOk := visitIdInterface.(float64)

	if !typeOk {
		return nil, 400
	}

	visitIdUint := uint(visitIdFloat)

	visitBytes := visitApiHandler.storage.GetVisitById(visitIdUint)

	if visitBytes != nil {
		return nil, 400
	}

	locationIdInterface, ok := newVisitMap["location"]

	if !ok {
		return nil, 400
	}

	userIdInterface, ok := newVisitMap["user"]

	if !ok {
		return nil, 400
	}

	VisitedAtInterface, ok := newVisitMap["visited_at"]

	if !ok {
		return nil, 400
	}

	markInterface, ok := newVisitMap["mark"]

	if !ok {
		return nil, 400
	}

	visit := new(entities.Visit)

	visit.Id = &visitIdUint

	locationIdFloat, typeOk := locationIdInterface.(float64)

	if !typeOk || locationIdFloat <= 0 {
		return nil, 400
	}

	locationIdUint := uint(locationIdFloat)

	location := visitApiHandler.storage.GetLocationById(locationIdUint)

	if location == nil {
		return nil, 400
	}

	visit.Location = &locationIdUint

	userIdFloat, typeOk := userIdInterface.(float64)

	if !typeOk || userIdFloat <= 0 {
		return nil, 400
	}

	userIdUint := uint(userIdFloat)

	user := visitApiHandler.storage.GetUserById(userIdUint)

	if user == nil {
		return nil, 400
	}

	visit.User = &userIdUint

	visitedAtFloat, typeOk := VisitedAtInterface.(float64)

	if !typeOk {
		return nil, 400
	}

	visitedAtInt := int(visitedAtFloat)

	visit.VisitedAt = &visitedAtInt

	markFloat, typeOk := markInterface.(float64)

	if !typeOk || (markFloat != 0 && markFloat != 1 && markFloat != 2 && markFloat != 3 && markFloat != 4 && markFloat != 5) {
		return nil, 400
	}

	markInt := int(markFloat)

	visit.Mark = &markInt

	visitApiHandler.storage.AddVisit(visit)
	visitApiHandler.storage.AddVisitByUserId(visit)
	visitApiHandler.storage.AddVisitByLocationId(visit)

	return []byte("{}"), 200
}
