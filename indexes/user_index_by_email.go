package indexes

import "sync"

type UserIndexByEmail struct {
	emails map[string]bool
	mutex  *sync.Mutex
}

func NewUserIndexByEmail() *UserIndexByEmail {
	return &UserIndexByEmail{emails: make(map[string]bool), mutex: new(sync.Mutex)}
}

func (userIndexByEmail *UserIndexByEmail) AddEmail(email string) {

	userIndexByEmail.mutex.Lock()

	userIndexByEmail.emails[email] = true

	userIndexByEmail.mutex.Unlock()
}

func (userIndexByEmail *UserIndexByEmail) IsEmailExist(email string) bool {

	_, isEmailExist := userIndexByEmail.emails[email]

	return isEmailExist
}

func (userIndexByEmail *UserIndexByEmail) DeleteEmail(email string) {

	userIndexByEmail.mutex.Lock()

	delete(userIndexByEmail.emails, email)

	userIndexByEmail.mutex.Unlock()
}
