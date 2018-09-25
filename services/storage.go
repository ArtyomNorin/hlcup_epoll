package services

import (
	"archive/zip"
	"fmt"
	"hlcup_epoll/entities"
	"hlcup_epoll/indexes"
	"io/ioutil"
	"log"
	"strings"
	"sync"
	"time"
	"sort"
	"github.com/json-iterator/go"
)

type Storage struct {
	errorLogger            *log.Logger
	infoLogger             *log.Logger
	userIndexByID          *indexes.UserIndexById
	locationIndexByID      *indexes.LocationIndexById
	visitIndexByID         *indexes.VisitIndexById
	userIndexByEmail       *indexes.UserIndexByEmail
	visitIndexByLocationID *indexes.VisitIndexByLocationId
	visitIndexByUserID     *indexes.VisitIndexByUserId
}

func NewStorage(errorLogger *log.Logger, infoLogger *log.Logger) *Storage {

	return &Storage{
		errorLogger:            errorLogger,
		infoLogger:             infoLogger,
		userIndexByID:          indexes.NewUserIndexById(),
		locationIndexByID:      indexes.NewLocationIndexById(),
		visitIndexByID:         indexes.NewVisitIndexById(),
		userIndexByEmail:       indexes.NewUserIndexByEmail(),
		visitIndexByLocationID: indexes.NewVisitIndexByLocationId(),
		visitIndexByUserID:     indexes.NewVisitIndexByUserId(),
	}
}

func (storage *Storage) Init(pathToArchive string, countConcurrentFiles int, waitGroup *sync.WaitGroup) {

	waitGroup.Add(countConcurrentFiles)

	channelOfZippedFiles := storage.readArchive(pathToArchive, countConcurrentFiles)

	for i := 0; i < countConcurrentFiles; i++ {

		go func() {

			for zippedJSONFile := range channelOfZippedFiles {

				startTime := time.Now()

				if strings.Contains(zippedJSONFile.Name, "user") {

					jsonOfFile := storage.extractJSONFromFile(zippedJSONFile)

					userCollection := new(entities.UserCollection)

					err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(jsonOfFile, userCollection)

					if err != nil {
						storage.errorLogger.Fatalln(err)
					}

					for _, user := range userCollection.Users {
						storage.AddUser(user)
					}
				}

				if strings.Contains(zippedJSONFile.Name, "location") {

					jsonOfFile := storage.extractJSONFromFile(zippedJSONFile)

					locationCollection := new(entities.LocationCollection)

					err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(jsonOfFile, locationCollection)

					if err != nil {
						storage.errorLogger.Fatalln(err)
					}

					for _, location := range locationCollection.Locations {
						storage.AddLocation(location)
					}
				}

				if strings.Contains(zippedJSONFile.Name, "visit") {

					jsonOfFile := storage.extractJSONFromFile(zippedJSONFile)

					visitCollection := new(entities.VisitCollection)

					err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(jsonOfFile, visitCollection)

					if err != nil {
						storage.errorLogger.Fatalln(err)
					}

					for _, visit := range visitCollection.Visits {
						storage.AddVisit(visit)
						storage.AddVisitByLocationId(visit)
						storage.AddVisitByUserId(visit)
					}
				}

				storage.infoLogger.Println(fmt.Sprintf("file %s is processed. Duration: %f", zippedJSONFile.Name, time.Now().Sub(startTime).Seconds()))
			}

			waitGroup.Done()
		}()

	}
}

func (storage *Storage) readArchive(pathToArchive string, countConcurrentFiles int) chan *zip.File {

	channelOfZippedFiles := make(chan *zip.File, countConcurrentFiles)

	zipReaderCloser, err := zip.OpenReader(pathToArchive)

	if err != nil {
		storage.errorLogger.Fatalln(err)
	}

	go func() {

		for _, zippedJSONFile := range zipReaderCloser.File {

			channelOfZippedFiles <- zippedJSONFile
		}

		close(channelOfZippedFiles)
	}()

	return channelOfZippedFiles
}

func (storage *Storage) extractJSONFromFile(zippedJSONFile *zip.File) []byte {

	readCloser, err := zippedJSONFile.Open()

	if err != nil {
		storage.errorLogger.Fatalln(err)
	}

	jsonOfFile, err := ioutil.ReadAll(readCloser)

	readCloser.Close()

	if err != nil {
		storage.errorLogger.Fatalln(err)
	}

	return jsonOfFile
}

func (storage *Storage) AddUser(user *entities.User) {

	err := storage.userIndexByID.AddUser(user)

	if err != nil {
		storage.errorLogger.Fatalln(err)
	}

	storage.userIndexByEmail.AddEmail(*user.Email)
}

func (storage *Storage) AddLocation(location *entities.Location) {

	err := storage.locationIndexByID.AddLocation(location)

	if err != nil {
		storage.errorLogger.Fatalln(err)
	}
}

func (storage *Storage) AddVisit(visit *entities.Visit) {

	err := storage.visitIndexByID.AddVisit(visit)

	if err != nil {
		storage.errorLogger.Fatalln(err)
	}
}

func (storage *Storage) AddVisitByUserId(visit *entities.Visit) {
	storage.visitIndexByUserID.AddVisit(visit)
}

func (storage *Storage) AddVisitByLocationId(visit *entities.Visit) {
	storage.visitIndexByLocationID.AddVisit(visit)
}

func (storage *Storage) GetUserById(userId uint) []byte {

	return storage.userIndexByID.GetUser(userId)
}

func (storage *Storage) GetLocationById(locationId uint) []byte {

	return storage.locationIndexByID.GetLocation(locationId)
}

func (storage *Storage) GetVisitById(visitId uint) []byte {

	return storage.visitIndexByID.GetVisit(visitId)
}

func (storage *Storage) DeleteVisitFromLocation(locationId uint, visitId uint) {

	storage.visitIndexByLocationID.DeleteVisit(locationId, visitId)
}

func (storage *Storage) DeleteVisitFromUser(userId uint, visitId uint) {

	storage.visitIndexByUserID.DeleteVisit(userId, visitId)
}

func (storage *Storage) DeleteEmail(email string) {

	storage.userIndexByEmail.DeleteEmail(email)
}

func (storage *Storage) IsEmailExist(email string) bool {

	return storage.userIndexByEmail.IsEmailExist(email)
}

//TODO need to refactor: logic mix
func (storage *Storage) GetVisitedPlacesByUser(visitFilter *VisitsFilter) *entities.VisitedPlaceCollection {

	userBytes := storage.GetUserById(*visitFilter.UserId)

	if userBytes == nil {
		return nil
	}

	visitsIds := storage.visitIndexByUserID.GetVisits(*visitFilter.UserId)

	visitedPlaceCollection := &entities.VisitedPlaceCollection{VisitedPlaces: make([]*entities.VisitedPlace, 0)}

	if len(visitsIds) == 0 {
		return visitedPlaceCollection
	}

	for _, visitId := range visitsIds {

		visit := new(entities.Visit)

		visitBytes := storage.GetVisitById(visitId)

		err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(visitBytes, visit)

		if err != nil {
			storage.errorLogger.Fatalln(err)
		}

		location := new(entities.Location)

		locationBytes := storage.GetLocationById(*visit.Location)

		err = jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(locationBytes, location)

		if err != nil {
			storage.errorLogger.Fatalln(err)
		}

		if
		!visitFilter.CheckFromDate(*visit.VisitedAt) ||
			!visitFilter.CheckToDate(*visit.VisitedAt) ||
			!visitFilter.CheckToDistance(*location.Distance) ||
			!visitFilter.CheckCountry(*location.Country) {
			continue
		}

		visitedPlace := &entities.VisitedPlace{VisitedAt: *visit.VisitedAt, Mark: *visit.Mark, Place: *location.Place}

		visitedPlaceCollection.VisitedPlaces = append(visitedPlaceCollection.VisitedPlaces, visitedPlace)
	}

	sort.Sort(visitedPlaceCollection)

	return visitedPlaceCollection
}

func (storage *Storage) GetVisitsByLocationId(locationId uint) []uint {

	return storage.visitIndexByLocationID.GetVisits(locationId)
}
