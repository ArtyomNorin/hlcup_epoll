package entities

type Location struct {
	Distance *uint   `json:"distance"`
	City     *string `json:"city"`
	Country  *string `json:"country"`
	Place    *string `json:"place"`
	Id       *uint   `json:"id"`
}

type LocationCollection struct {
	Locations []*Location `json:"locations"`
}
