package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
)

type UserRepo struct {
	cli    *api.Client
	logger *log.Logger
}

const (
	users = "users/%s"
	all   = "users"
)

func generateKey(email string) (string, string) {
	id := email
	return fmt.Sprintf(users, id), id
}

func New(logger *log.Logger) (*UserRepo, error) {
	config := api.DefaultConfig()
	config.Address = fmt.Sprintf("%s:%s", "consul", "8500")
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &UserRepo{
		cli:    client,
		logger: logger,
	}, nil
}

func (ur *UserRepo) Insert(user *User) error {
	log.Println("Usli u Insert")
	kv := ur.cli.KV()

	//dbId, id := generateKey()
	user.ID = uuid.New().String();

	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	userKeyValue := &api.KVPair{Key: user.Email, Value: data}
	_, err = kv.Put(userKeyValue, nil)
	if err != nil {
		return err
	}

	return nil
}

func (pr *UserRepo) GetAll() (Users, error) {
	kv := pr.cli.KV()
	data, _, err := kv.List(all, nil)
	if err != nil {
		return nil, err
	}

	users := Users{}
	for _, pair := range data {
		user := &User{}
		err = json.Unmarshal(pair.Value, user)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}
func (ur *UserRepo) Get(email string) (*User, error) {
	kv := ur.cli.KV()

	pair, _, err := kv.Get(constructKey(email), nil)
	if err != nil {
		return nil, err
	}
	// If pair is nil -> no object found for given id -> return nil
	if pair == nil {
		log.Println("blabla:",pair)
		return nil, nil
	}

	user := &User{}
	err = json.Unmarshal(pair.Value, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}
func constructKey(email string) string {
	return fmt.Sprintf(users, email)
}