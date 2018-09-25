package entities

type Visit struct {
	Mark      *int  `json:"mark, omitempty"`
	VisitedAt *int  `json:"visited_at, omitempty"`
	User      *uint `json:"user, omitempty"`
	Id        *uint `json:"id, omitempty"`
	Location  *uint `json:"location, omitempty"`
}

type VisitCollection struct {
	Visits []*Visit `json:"visits"`
}