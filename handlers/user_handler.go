package handlers

import (
	"hlcup_epoll/services"
	"log"
		"github.com/asaskevich/govalidator"
	"strconv"
	"time"
	"os"
	"bufio"
	"encoding/json"
		"github.com/json-iterator/go"
	"hlcup_epoll/entities"
	"net/http"
)

type UserApiHandler struct {
	storage            *services.Storage
	errLogger          *log.Logger
	infoLogger         *log.Logger
	timeDataGeneration time.Time
}

func NewUserApiHandler(storage *services.Storage, errLogger *log.Logger, infoLogger *log.Logger, pathToOptions string) *UserApiHandler {

	file, _ := os.Open(pathToOptions)

	fileScanner := bufio.NewScanner(file)

	fileScanner.Scan()

	timeDataGeneration, err := strconv.Atoi(fileScanner.Text())

	if err != nil {
		errLogger.Fatalln(err)
	}

	return &UserApiHandler{storage: storage, errLogger: errLogger, infoLogger: infoLogger, timeDataGeneration: time.Unix(int64(timeDataGeneration), 0)}
}

func (userApiHandler *UserApiHandler) GetById(request *http.Request, userIdString string) ([]byte, int) {

	userId, err := strconv.Atoi(userIdString)

	if err != nil {
		return nil, 404
	}

	if userId <= 0 {
		return nil, 404
	}

	user := userApiHandler.storage.GetUserById(uint(userId))

	if user == nil {
		return nil, 404
	} else {
		return user, 200
	}
}

func (userApiHandler *UserApiHandler) GetVisitedPlaces(request *http.Request, userIdString string) ([]byte, int) {

	userId, err := strconv.Atoi(userIdString)

	if err != nil {
		return nil, 404
	}

	if userId <= 0 {
		return nil, 404
	}

	filter := services.InitVisitFilter(userApiHandler.timeDataGeneration)

	userIdUint := uint(userId)

	filter.UserId = &userIdUint

	if value, ok := request.URL.Query()["fromDate"]; ok {

		fromDateString := string(value[0])

		if !govalidator.IsNumeric(fromDateString) || fromDateString == "" {
			return nil, 400
		}

		fromDateInt, err := strconv.Atoi(fromDateString)

		if err != nil {
			userApiHandler.errLogger.Println(err)
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
			userApiHandler.errLogger.Println(err)
		}

		filter.ToDate = &toDateInt
	}

	if value, ok := request.URL.Query()["toDistance"]; ok {

		toDistanceString := string(value[0])

		if !govalidator.IsNumeric(toDistanceString) || toDistanceString == "" {
			return nil, 400
		}

		toDistanceInt, err := strconv.Atoi(toDistanceString)

		if err != nil {
			userApiHandler.errLogger.Println(err)
		}

		toDistanceUint := uint(toDistanceInt)

		filter.ToDistance = &toDistanceUint
	}

	if value, ok := request.URL.Query()["country"]; ok {

		country := string(value[0])

		if country == "" || len(country) > 50 {
			return nil, 400
		}

		filter.Country = &country
	}

	visitedPlaceCollection := userApiHandler.storage.GetVisitedPlacesByUser(filter)

	if visitedPlaceCollection == nil {
		return nil, 404
	}

	visitedPlaceCollectionBytes, err := json.Marshal(visitedPlaceCollection)

	if err != nil {
		userApiHandler.errLogger.Fatalln(err)
	}

	return visitedPlaceCollectionBytes, 200
}

func (userApiHandler *UserApiHandler) Update(request *http.Request, userIdString string) ([]byte, int) {

	userId, err := strconv.Atoi(userIdString)

	if err != nil {
		return nil, 404
	}

	userBytes := userApiHandler.storage.GetUserById(uint(userId))

	if userBytes == nil {
		return nil, 404
	}

	newUserMap := make(map[string]interface{})

	err = jsoniter.NewDecoder(request.Body).Decode(&newUserMap)

	if err != nil {
		return nil, 400
	}

	user := new(entities.User)

	err = jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(userBytes, user)

	if err != nil {
		userApiHandler.errLogger.Fatalln(err)
	}

	if value, ok := newUserMap["email"]; ok {

		email, typeOk := value.(string)

		if value == nil || !typeOk || len(email) > 100 {
			return nil, 400
		}

		user.Email = &email
	}

	if value, ok := newUserMap["first_name"]; ok {

		firstName, typeOk := value.(string)

		if value == nil || !typeOk || len(firstName) > 50 {
			return nil, 400
		}

		user.FirstName = &firstName
	}

	if value, ok := newUserMap["last_name"]; ok {

		lastName, typeOk := value.(string)

		if value == nil || !typeOk || len(lastName) > 50 {
			return nil, 400
		}

		user.LastName = &lastName
	}

	if value, ok := newUserMap["gender"]; ok {

		gender, typeOk := value.(string)

		if value == nil || !typeOk || (gender != "m" && gender != "f") {
			return nil, 400
		}

		user.Gender = &gender
	}

	if value, ok := newUserMap["birth_date"]; ok {

		birthDateFloat, typeOk := value.(float64)

		if value == nil || !typeOk {
			return nil, 400
		}

		birthDate := int(birthDateFloat)

		user.BirthDate = &birthDate
	}

	userApiHandler.storage.AddUser(user)

	return []byte("{}"), 200
}

func (userApiHandler *UserApiHandler) Create(request *http.Request) ([]byte, int) {

	newUserMap := make(map[string]interface{})

	err := jsoniter.NewDecoder(request.Body).Decode(&newUserMap)

	if err != nil {
		return nil, 400
	}

	userIdInterface, ok := newUserMap["id"]

	if !ok {
		return nil, 400
	}

	userIdFloat, typeOk := userIdInterface.(float64)

	if !typeOk {
		return nil, 400
	}

	userIdUint := uint(userIdFloat)

	userBytes := userApiHandler.storage.GetLocationById(userIdUint)

	if userBytes != nil {
		return nil, 400
	}

	emailInterface, ok := newUserMap["email"]

	if !ok {
		return nil, 400
	}

	email, typeOk := emailInterface.(string)

	if !typeOk || len(email) > 100 {
		return nil, 400
	}

	isEmailExist := userApiHandler.storage.IsEmailExist(email)

	if isEmailExist {
		return nil, 400
	}

	firstNameInterface, ok := newUserMap["first_name"]

	if !ok {
		return nil, 400
	}

	lastNameInterface, ok := newUserMap["last_name"]

	if !ok {
		return nil, 400
	}

	genderInterface, ok := newUserMap["gender"]

	if !ok {
		return nil, 400
	}

	birthDateInterface, ok := newUserMap["birth_date"]

	if !ok {
		return nil, 400
	}

	user := new(entities.User)

	user.Id = &userIdUint

	user.Email = &email

	firstName, typeOk := firstNameInterface.(string)

	if !typeOk || len(firstName) > 50 {
		return nil, 400
	}

	user.FirstName = &firstName

	lastName, typeOk := lastNameInterface.(string)

	if !typeOk || len(lastName) > 50 {
		return nil, 400
	}

	user.LastName = &lastName

	gender, typeOk := genderInterface.(string)

	if !typeOk || (gender != "m" && gender != "f") {
		return nil, 400
	}

	user.Gender = &gender

	birthDateFloat, typeOk := birthDateInterface.(float64)

	if !typeOk {
		return nil, 400
	}

	birthDate := int(birthDateFloat)

	user.BirthDate = &birthDate

	userApiHandler.storage.AddUser(user)

	return []byte("{}"), 200
}