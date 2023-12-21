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

func generateKey() (string, string) {
	id := uuid.New().String()
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

	dbId, id := generateKey()
	user.ID = id

	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	userKeyValue := &api.KVPair{Key: dbId, Value: data}
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
