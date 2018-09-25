package indexes

import (
	"hlcup_epoll/entities"
	"sync"
	"github.com/json-iterator/go"
)

type UserIndexById struct {
	users map[uint][]byte
	mutex *sync.Mutex
}

func NewUserIndexById() *UserIndexById {
	return &UserIndexById{users: make(map[uint][]byte), mutex: new(sync.Mutex)}
}

func (userIndexById *UserIndexById) AddUser(user *entities.User) error {

	encodedUser, err := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(user)

	if err != nil {
		return err
	}

	userIndexById.mutex.Lock()

	userIndexById.users[*user.Id] = encodedUser

	userIndexById.mutex.Unlock()

	return nil
}

func (userIndexById *UserIndexById) GetUser(userId uint) []byte {

	userBytes, isIdExist := userIndexById.users[userId]

	if !isIdExist {
		return nil
	}

	return userBytes
}
