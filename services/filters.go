package services

import (
	"strings"
	"time"
)

type Int *int
type Uint *uint
type String *string

type VisitsFilter struct {
	timeDataGeneration time.Time
	UserId             Uint
	FromDate           Int
	ToDate             Int
	Country            String
	ToDistance         Uint
	FromAge            Int
	ToAge              Int
	Gender             String
	LocationId         Uint
}

func InitVisitFilter(timeDataGeneration time.Time) *VisitsFilter {
	return &VisitsFilter{timeDataGeneration: timeDataGeneration}
}

func (visitFilter *VisitsFilter) CheckCountry(country string) bool {

	if visitFilter.Country == nil {
		return true
	}

	return strings.ToLower(country) == strings.ToLower(*visitFilter.Country)
}

func (visitFilter *VisitsFilter) CheckFromDate(visitedAt int) bool {

	if visitFilter.FromDate == nil {
		return true
	}

	return visitedAt > *visitFilter.FromDate
}

func (visitFilter *VisitsFilter) CheckToDate(visitedAt int) bool {

	if visitFilter.ToDate == nil {
		return true
	}

	return visitedAt < *visitFilter.ToDate
}

func (visitFilter *VisitsFilter) CheckToDistance(toDistance uint) bool {

	if visitFilter.ToDistance == nil {
		return true
	}

	return toDistance < *visitFilter.ToDistance
}

func (visitFilter *VisitsFilter) CheckFromAge(birthDate int) bool {

	if visitFilter.FromAge == nil {
		return true
	}

	return int(visitFilter.timeDataGeneration.AddDate(-*visitFilter.FromAge, 0, 0).Unix()) > birthDate
}

func (visitFilter *VisitsFilter) CheckToAge(birthDate int) bool {

	if visitFilter.ToAge == nil {
		return true
	}

	return int(visitFilter.timeDataGeneration.AddDate(-*visitFilter.ToAge, 0, 0).Unix()) < birthDate
}

func (visitFilter *VisitsFilter) CheckGender(gender string) bool {

	if visitFilter.Gender == nil {
		return true
	}

	return gender == *visitFilter.Gender
}
