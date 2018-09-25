package handlers

import (
	"hlcup_epoll/services"
	"log"
	"github.com/asaskevich/govalidator"
	"strconv"
	"os"
	"bufio"
	"time"
	"hlcup_epoll/entities"
	"math"
	"github.com/json-iterator/go"
	"net/http"
)

type LocationApiHandler struct {
	storage            *services.Storage
	errLogger          *log.Logger
	infoLogger         *log.Logger
	timeDataGeneration time.Time
}

func NewLocationApiHandler(storage *services.Storage, errLogger *log.Logger, infoLogger *log.Logger, pathToOptions string) *LocationApiHandler {

	file, _ := os.Open(pathToOptions)

	fileScanner := bufio.NewScanner(file)

	fileScanner.Scan()

	timeDataGeneration, err := strconv.Atoi(fileScanner.Text())

	if err != nil {
		errLogger.Fatalln(err)
	}

	return &LocationApiHandler{storage: storage, errLogger: errLogger, infoLogger: infoLogger, timeDataGeneration: time.Unix(int64(timeDataGeneration), 0)}
}

func (locationApiHandler *LocationApiHandler) GetById(request *http.Request, locationIdString string) ([]byte, int) {

	locationId, err := strconv.Atoi(locationIdString)

	if err != nil {
		return nil, 404
	}

	location := locationApiHandler.storage.GetLocationById(uint(locationId))

	if location == nil {
		return nil, 404
	} else {
		return location, 200
	}
}

func (locationApiHandler *LocationApiHandler) GetAverageMark(request *http.Request, locationIdString string) ([]byte, int) {

	locationIdInt, err := strconv.Atoi(locationIdString)

	if err != nil {
		return nil, 404
	}

	filter := services.InitVisitFilter(locationApiHandler.timeDataGeneration)

	locationIdUint := uint(locationIdInt)

	filter.LocationId = &locationIdUint

	if value, ok := request.URL.Query()["fromDate"]; ok {

		fromDateString := string(value[0])

		if !govalidator.IsNumeric(fromDateString) || fromDateString == "" {
			return nil, 400
		}

		fromDateInt, err := strconv.Atoi(fromDateString)

		if err != nil {
			locationApiHandler.errLogger.Println(err)
		}

		filter.FromDate = &fromDateInt
	}

	if value, ok := request.URL.Query()["toDate"]; ok {

		toDateString := string(value[0])

		if !govalidator.IsNumeric(toDateString) || toDateString == "" {
			return nil, 400
		}

		toDateInt, err := strconv.Atoi(toDateString)

		if err != nil {
			locationApiHandler.errLogger.Println(err)
		}

		filter.ToDate = &toDateInt
	}

	if value, ok := request.URL.Query()["fromAge"]; ok {

		fromAgeString := string(value[0])

		if !govalidator.IsNumeric(fromAgeString) || fromAgeString == "" {
			return nil, 400
		}

		fromAgeInt, err := strconv.Atoi(fromAgeString)

		if err != nil {
			locationApiHandler.errLogger.Println(err)
		}

		filter.FromAge = &fromAgeInt
	}

	if value, ok := request.URL.Query()["toAge"]; ok {

		toAgeString := string(value[0])

		if !govalidator.IsNumeric(toAgeString) || toAgeString == "" {
			return nil, 400
		}

		toAgeInt, err := strconv.Atoi(toAgeString)

		if err != nil {
			locationApiHandler.errLogger.Println(err)
		}

		filter.ToAge = &toAgeInt
	}

	if value, ok := request.URL.Query()["gender"]; ok {

		gender := string(value[0])

		if gender == "" || (gender != "m" && gender != "f") {
			return nil, 400
		}

		filter.Gender = &gender
	}

	locationBytes := locationApiHandler.storage.GetLocationById(*filter.LocationId)

	if locationBytes == nil {
		return nil, 404
	}

	visitsIds := locationApiHandler.storage.GetVisitsByLocationId(*filter.LocationId)

	visitCollection := new(entities.VisitCollection)
	sumOfMarks := 0

	for _, visitId := range visitsIds {

		visitBytes := locationApiHandler.storage.GetVisitById(visitId)

		visit := new(entities.Visit)

		err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(visitBytes, visit)

		if err != nil {
			locationApiHandler.errLogger.Fatalln(err)
		}

		userBytes := locationApiHandler.storage.GetUserById(*visit.User)

		user := new(entities.User)

		err = jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(userBytes, user)

		if err != nil {
			locationApiHandler.errLogger.Fatalln(err)
		}

		if !filter.CheckFromAge(*user.BirthDate) ||
			!filter.CheckToAge(*user.BirthDate) ||
			!filter.CheckToDate(*visit.VisitedAt) ||
			!filter.CheckFromDate(*visit.VisitedAt) ||
			!filter.CheckGender(*user.Gender) {
			continue
		}

		sumOfMarks += *visit.Mark
		visitCollection.Visits = append(visitCollection.Visits, visit)
	}

	locationAvgMark := &entities.LocationAvgMark{Avg: 0}

	if len(visitCollection.Visits) != 0 {
		locationAvgMark.Avg = math.Round(float64(sumOfMarks)/float64(len(visitCollection.Visits))*100000) / 100000
	}

	locationAvgMarkBytes, err := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(locationAvgMark)

	if err != nil {
		locationApiHandler.errLogger.Fatalln(err)
	}

	return locationAvgMarkBytes, 200
}

func (locationApiHandler *LocationApiHandler) Update(request *http.Request, locationIdString string) ([]byte, int) {

	locationId, err := strconv.Atoi(locationIdString)

	if err != nil {
		return nil, 404
	}

	locationBytes := locationApiHandler.storage.GetLocationById(uint(locationId))

	if locationBytes == nil {
		return nil, 404
	}

	newLocationMap := make(map[string]interface{})

	err = jsoniter.NewDecoder(request.Body).Decode(&newLocationMap)

	if err != nil {
		return nil, 400
	}

	location := new(entities.Location)

	err = jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(locationBytes, location)

	if err != nil {
		locationApiHandler.errLogger.Fatalln(err)
	}

	if value, ok := newLocationMap["place"]; ok {

		place, typeOk := value.(string)

		if value == nil || !typeOk {
			return nil, 400
		}

		location.Place = &place
	}

	if value, ok := newLocationMap["country"]; ok {

		country, typeOk := value.(string)

		if value == nil || !typeOk || len(country) > 50 {
			return nil, 400
		}

		location.Country = &country
	}

	if value, ok := newLocationMap["city"]; ok {

		city, typeOk := value.(string)

		if value == nil || !typeOk || len(city) > 50 {
			return nil, 400
		}

		location.City = &city
	}

	if value, ok := newLocationMap["distance"]; ok {

		distance, typeOk := value.(float64)

		if value == nil || !typeOk || distance <= 0 {
			return nil, 400
		}

		distanceAsUint := uint(distance)

		location.Distance = &distanceAsUint
	}

	locationApiHandler.storage.AddLocation(location)

	return []byte("{}"), 200
}

func (locationApiHandler *LocationApiHandler) Create(request *http.Request) ([]byte, int) {

	newLocationMap := make(map[string]interface{})

	err := jsoniter.NewDecoder(request.Body).Decode(&newLocationMap)

	if err != nil {
		return nil, 400
	}

	locationIdInterface, ok := newLocationMap["id"]

	if !ok {
		return nil, 400
	}

	locationIdFloat, typeOk := locationIdInterface.(float64)

	if !typeOk {
		return nil, 400
	}

	locationIdUint := uint(locationIdFloat)

	locationBytes := locationApiHandler.storage.GetLocationById(locationIdUint)

	if locationBytes != nil {
		return nil, 400
	}

	placeInterface, ok := newLocationMap["place"]

	if !ok {
		return nil, 400
	}

	countryInterface, ok := newLocationMap["country"]

	if !ok {
		return nil, 400
	}

	cityInterface, ok := newLocationMap["city"]

	if !ok {
		return nil, 400
	}

	distanceInterface, ok := newLocationMap["distance"]

	if !ok {
		return nil, 400
	}

	location := new(entities.Location)

	location.Id = &locationIdUint

	place, typeOk := placeInterface.(string)

	if !typeOk {
		return nil, 400
	}

	location.Place = &place

	country, typeOk := countryInterface.(string)

	if !typeOk || len(country) > 50 {
		return nil, 400
	}

	location.Country = &country

	city, typeOk := cityInterface.(string)

	if !typeOk || len(city) > 50 {
		return nil, 400
	}

	location.City = &city

	distanceFloat, typeOk := distanceInterface.(float64)

	if !typeOk || distanceFloat <= 0 {
		return nil, 400
	}

	distance := uint(distanceFloat)

	location.Distance = &distance

	locationApiHandler.storage.AddLocation(location)

	return []byte("{}"), 200
}
