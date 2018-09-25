package entities

type VisitedPlace struct {
	Mark      int    `json:"mark"`
	VisitedAt int    `json:"visited_at"`
	Place     string `json:"place"`
}

type VisitedPlaceCollection struct {
	VisitedPlaces []*VisitedPlace `json:"visits"`
}

func (visitedPlaceCollection *VisitedPlaceCollection) Len() int {
	return len(visitedPlaceCollection.VisitedPlaces)
}

func (visitedPlaceCollection *VisitedPlaceCollection) Swap(i, j int) {
	visitedPlaceCollection.VisitedPlaces[i], visitedPlaceCollection.VisitedPlaces[j] = visitedPlaceCollection.VisitedPlaces[j], visitedPlaceCollection.VisitedPlaces[i]
}

func (visitedPlaceCollection *VisitedPlaceCollection) Less(i, j int) bool {
	return visitedPlaceCollection.VisitedPlaces[i].VisitedAt < visitedPlaceCollection.VisitedPlaces[j].VisitedAt
}