package models

import (
	"fmt"

	"github.com/go-redis/redis"
	"github.com/nomnom-ray/webGCS/util"
)

//User key --
type User struct {
	id int64
}

//NewUser constructor
func NewUser(username string) (*User, error) {
	id, err := client.Incr("user:next-id").Result()
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("user:%d", id)

	pipe := client.Pipeline()
	pipe.HSet(key, "id", id)
	pipe.HSet(key, "username", username)
	pipe.HSet("user:by-username", username, id)
	_, err = pipe.Exec()
	if err != nil {
		return nil, err
	}
	return &User{id}, nil
}

//RegisterUser --
func RegisterUser(username string) error {
	_, err := NewUser(username)
	return err
}

//GetUserbyName returns a redis key based on username
func GetUserbyName(username string) (*User, error) {
	id, err := client.HGet("user:by-username", username).Int64()
	if err == redis.Nil {
		return nil, util.ErrUserNotFound
	} else if err != nil {
		return nil, err
	}
	return &User{id}, nil
}

//GetID gets user attribute from redis key
func (user *User) GetID() (int64, error) {
	return user.id, nil
}

func CheckUserExist(username string) (bool, error) {
	exists, err := client.HExists("user:by-username", username).Result()
	if err != nil {
		return false, err
	}
	return exists, nil
}
