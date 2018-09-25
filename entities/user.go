package entities

type User struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Gender    *string `json:"gender"`
	Email     *string `json:"email"`
	BirthDate *int    `json:"birth_date"`
	Id        *uint   `json:"id"`
}

type UserCollection struct {
	Users []*User `json:"users"`
}
