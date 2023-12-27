package main

import (
	"encoding/json"
	"fmt"
	"log"

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

func generateKey(Id string) string {
	id := Id
	return fmt.Sprintf(users, id)
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

	dbId := generateKey(user.ID)

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

func (ur *UserRepo) Get(id string) (*ResponseUser, error) {
	kv := ur.cli.KV()

	pair, _, err := kv.Get(constructKey(id), nil)
	if err != nil {
		return nil, err
	}
	if pair == nil {
		log.Println("blabla:", pair)
		return nil, nil
	}

	user := &ResponseUser{}
	err = json.Unmarshal(pair.Value, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func constructKey(id string) string {
	return fmt.Sprintf(users, id)
}

func (ur *UserRepo) UpdateUser(user *User) error {
	kv := ur.cli.KV()

	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	productKeyValue := &api.KVPair{Key: constructKey(user.ID), Value: data}
	_, err = kv.Put(productKeyValue, nil)
	if err != nil {
		return err
	}

	return nil
}
