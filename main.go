package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	totalRequest    = 3
	usersPerRequest = 5000
	outputFile      = "users_cache.json"
)

type User struct {
	Gender    string `json:"gender"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	City      string `json:"city"`
	Country   string `json:"country"`
	UUID      string `json:"uuid"`
}

type apiResponse struct {
	Results []struct {
		Gender string `json:"gender"`
		Name   struct {
			First string `json:"first"`
			Last  string `json:"last"`
		} `json:"name"`
		Email    string `json:"email"`
		Location struct {
			City    string `json:"city"`
			Country string `json:"country"`
		} `json:"location"`
		Login struct {
			UUID string `json:"uuid"`
		} `json:"login"`
	} `json:"results"`
}

type FullResponse struct {
	Users         []User `json:"users"`
	MaleCount     int    `json:"male_count"`
	FemaleCount   int    `json:"female_count"`
	TotalUsers    int    `json:"total_users"`
	ExecutionTime string `json:"execution_time"`
}

func main() {
	log.Println("Starting server on http://localhost:8080")
	http.HandleFunc("/users", getUsersHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func getUsersHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if fileExists(outputFile) {
		log.Println("Cache found, loading users from file...")
		cachedUsers, err := loadUsersFromFile()
		if err == nil {
			elapsed := time.Since(start)
			log.Printf("Served %d cached users in %.2f seconds", len(cachedUsers), elapsed.Seconds())
			serveWithStats(w, cachedUsers, elapsed)
			return
		}
		log.Println("Error loading cache, fetching new data...")
	}

	log.Printf("Fetching %d requests, each for %d users...", totalRequest, usersPerRequest)

	var wg sync.WaitGroup
	userChannel := make(chan []User, totalRequest)

	for i := 1; i <= totalRequest; i++ {
		wg.Add(1)
		go func(requestNum int) {
			defer wg.Done()
			log.Printf("Request #%d started", requestNum)
			users, err := fetchUsers(usersPerRequest)
			if err != nil {
				log.Printf("Request #%d failed: %v", requestNum, err)
				return
			}
			userChannel <- users
		}(i)
	}

	go func() {
		wg.Wait()
		close(userChannel)
	}()

	var allUsers []User
	for batch := range userChannel {
		allUsers = append(allUsers, batch...)
	}

	elapsed := time.Since(start)
	log.Printf("Finished collecting %d users in %.2f seconds", len(allUsers), elapsed.Seconds())

	saveUsersToFile(allUsers)
	serveWithStats(w, allUsers, elapsed)
}

func fetchUsers(count int) ([]User, error) {
	url := fmt.Sprintf("https://randomuser.me/api/?results=%d&inc=gender,name,email,location,login&noinfo", count)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error on GET: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-OK response: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %v", err)
	}

	var apiResp apiResponse
	err = json.Unmarshal(body, &apiResp)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	var users []User
	for _, item := range apiResp.Results {
		users = append(users, User{
			Gender:    item.Gender,
			FirstName: item.Name.First,
			LastName:  item.Name.Last,
			Email:     item.Email,
			City:      item.Location.City,
			Country:   item.Location.Country,
			UUID:      item.Login.UUID,
		})
	}
	return users, nil
}

func saveUsersToFile(users []User) {
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		log.Printf("Error marshaling users: %v", err)
		return
	}
	err = os.WriteFile(outputFile, data, 0644)
	if err != nil {
		log.Printf("Error writing to file: %v", err)
	}
}

func loadUsersFromFile() ([]User, error) {
	data, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, err
	}
	var users []User
	err = json.Unmarshal(data, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func serveWithStats(w http.ResponseWriter, users []User, duration time.Duration) {
	var maleCount, femaleCount int
	for _, u := range users {
		if u.Gender == "male" {
			maleCount++
		} else if u.Gender == "female" {
			femaleCount++
		}
	}

	response := FullResponse{
		Users:         users,
		MaleCount:     maleCount,
		FemaleCount:   femaleCount,
		TotalUsers:    len(users),
		ExecutionTime: fmt.Sprintf("%.2f seconds", duration.Seconds()),
	}

	log.Printf("Served %d users with stats in %.2f seconds", len(users), duration.Seconds())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
