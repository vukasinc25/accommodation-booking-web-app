package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/consul/api"
)

type UserRepo struct {
	cli                        *api.Client
	logger                     *log.Logger
	reservation_service_string string
}

const (
	users          = "users/%s"
	all            = "users"
	hostGrades     = "hostGrades/%s"
	hostGradeIndex = "hostGradeIndex/%s"
)

func generateKey(Id string) string {
	id := Id
	return fmt.Sprintf(users, id)
}

func New(logger *log.Logger, conn_reservation_service_address string) (*UserRepo, error) {
	config := api.DefaultConfig()
	config.Address = fmt.Sprintf("%s:%s", "consul", "8500")
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &UserRepo{
		cli:                        client,
		logger:                     logger,
		reservation_service_string: conn_reservation_service_address,
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

func (ur *UserRepo) Delete(id string) error {
	kv := ur.cli.KV()

	_, err := kv.Delete(constructKey(id), nil)
	if err != nil {
		return err
	}

	return nil
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

func (ur *UserRepo) CreateHostGrade(hostGrade *HostGrade) error {
	log.Println("Usli u CreateHostGrade")
	log.Println("HostGrade:", hostGrade)
	kv := ur.cli.KV()

	dbId := fmt.Sprintf(hostGrades, hostGrade.ID)

	data, err := json.Marshal(hostGrade)
	if err != nil {
		return err
	}

	log.Println("Data:", data)

	KeyValue := &api.KVPair{Key: dbId, Value: data}
	_, err = kv.Put(KeyValue, nil)
	if err != nil {
		return err
	}

	// Add the HostGrade ID to the index
	indexKey := fmt.Sprintf(hostGradeIndex, hostGrade.HostId)
	indexValue, _, err := kv.Get(indexKey, nil)
	if err != nil {
		return err
	}

	log.Println("Index Value:", indexValue)

	var ids []string
	if indexValue != nil {
		if err := json.Unmarshal(indexValue.Value, &ids); err != nil {
			return err
		}
	}

	// Add the new ID to the list
	ids = append(ids, hostGrade.ID)

	// Marshal the list of IDs into a JSON array
	jsonArray, err := json.Marshal(ids)
	if err != nil {
		return err
	}

	indexValue = &api.KVPair{Key: indexKey, Value: jsonArray}
	_, err = kv.Put(indexValue, nil)
	if err != nil {
		return err
	}

	return nil
}

func (ur *UserRepo) DeleteHostGrade(id string, userId string) error {
	kv := ur.cli.KV()

	log.Println("Ovde1")
	// Retrieve HostGrade by ID to get the HostId
	hostGradeKey := fmt.Sprintf(hostGrades, strings.TrimSpace(id))
	log.Println("HostGradeKey:", hostGradeKey)
	pair, _, err := kv.Get(hostGradeKey, nil)
	if err != nil {
		log.Println("Error in getting host grade by id:", err)
		return err
	}
	log.Println("Ovde2")

	if pair == nil {
		return errors.New("HostGrade not found")
	}

	var hg HostGrade
	if err := json.Unmarshal(pair.Value, &hg); err != nil {
		return err
	}

	if hg.UserId != strings.Trim(userId, `"`) {
		return errors.New("cant delete host grade")
	}

	log.Println("Ovde4")

	// Log initial values
	log.Printf("Deleting HostGrade: %s, HostId: %s\n", id, hg.HostId)

	// Delete the HostGrade
	_, err = kv.Delete(hostGradeKey, nil)
	if err != nil {
		return err
	}

	log.Println("HostGrade deleted successfully")

	// Remove the ID from the hostGradeIndex table
	indexKey := fmt.Sprintf(hostGradeIndex, hg.HostId)
	indexValue, _, err := kv.Get(indexKey, nil)
	if err != nil {
		return err
	}

	if indexValue != nil {
		log.Printf("Original indexValue: %s\n", indexValue.Value)

		updatedIndexValue := removeFromCSV(string(indexValue.Value), id)
		log.Printf("Updated indexValue: %s\n", updatedIndexValue)

		_, err := kv.Put(&api.KVPair{Key: indexKey, Value: []byte(updatedIndexValue)}, nil)
		if err != nil {
			return err
		}

		log.Println("HostGradeIndex updated successfully")
	} else {
		log.Println("HostGradeIndex not found")
	}

	return nil
}

func removeFromCSV(csvString, idToRemove string) string {
	log.Printf("csvString: %q\n", csvString)
	log.Printf("idToRemove: %q\n", idToRemove)
	var ids []string
	if err := json.Unmarshal([]byte(csvString), &ids); err != nil {
		log.Println("Error unmarshaling CSV string:", err)
	}

	var updatedIDs []string
	log.Println("ids", ids)
	log.Println("updatedIDs1", updatedIDs)
	for _, existingID := range ids {
		log.Println("Usli u for")
		existingID = strings.TrimSpace(existingID)
		idToRemove = strings.TrimSpace(idToRemove)
		if existingID != idToRemove {
			log.Println("existingID", existingID)
			updatedIDs = append(updatedIDs, existingID)
			log.Println("updatedIDs2", updatedIDs)
		}
	}
	log.Println("FinalUpdatedIDs", updatedIDs)
	jsonArrayString, err := json.Marshal(updatedIDs)
	if err != nil {
		log.Println("Error marshaling JSON array:", err)
		return ""
	}
	log.Println("FinalUpdatedIDs2", strings.Join(updatedIDs, ","))

	return string(jsonArrayString)
}

func (ur *UserRepo) GetAllReservatinsForUserByHostId(userId string, hostId string) (*http.Response, error) {
	url := ur.reservation_service_string + "/api/reservations/by_user_for_host_id/" + userId + "/" + hostId

	log.Println("Url", url)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return httpResp, nil
}

func (ur *UserRepo) GetAllHostGradesByHostId(hostId string) ([]HostGrade, error) {
	log.Println("Usli u GetAllHostGradesByHostId")
	kv := ur.cli.KV()

	indexKey := fmt.Sprintf(hostGradeIndex, strings.Trim(hostId, `"`))
	indexValue, _, err := kv.Get(indexKey, nil)
	if err != nil {
		return nil, err
	}

	log.Println("IndeKey", indexKey)
	log.Println("IndexValie", indexValue)

	if indexValue == nil {
		return nil, nil // No HostGrades found for the specified HostId
	}

	log.Println("Index value:", indexValue)

	var hostGradeIDs []string
	if err := json.Unmarshal(indexValue.Value, &hostGradeIDs); err != nil {
		return nil, err
	}
	log.Println("hostGradeIDs:", hostGradeIDs)
	hostGrades := make([]HostGrade, 0)
	log.Println("hostGrades:", hostGrades)

	for _, id := range hostGradeIDs {
		log.Println("hostGradeID:", id)
		hostGrade, err := ur.GetHostGradeByID(id)
		if err != nil {
			return nil, err
		}
		log.Println("hostGrade:", hostGrade)

		hostGrades = append(hostGrades, *hostGrade)
	}

	return hostGrades, nil
}

// GetHostGradeByID retrieves a host grade by its ID
func (ur *UserRepo) GetHostGradeByID(id string) (*HostGrade, error) {
	kv := ur.cli.KV()

	hostGradeKey := fmt.Sprintf(hostGrades, id)
	pair, _, err := kv.Get(hostGradeKey, nil)
	if err != nil {
		return nil, err
	}

	if pair == nil {
		return nil, errors.New("HostGrade not found")
	}

	log.Println("HostGrade:", pair)

	var hg HostGrade
	if err := json.Unmarshal(pair.Value, &hg); err != nil {
		return nil, err
	}

	return &hg, nil
}
