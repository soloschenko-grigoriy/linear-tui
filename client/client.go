package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Issue struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	State struct {
		Name string `json:"name"`
	} `json:"state"`
	Priority int `json:"priority"`
}

type Response struct {
	Data struct {
		Issues struct {
			Nodes []Issue `json:"nodes"`
		} `json:"issues"`
	} `json:"data"`
}

func FetchIssues() ([]Issue, error) {
	key := os.Getenv("LINEAR_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("LINEAR_API_KEY not set")
	}

	query := `{"query": "{ issues(first: 20) { nodes { id title state { name } priority } } }" }`

	body := bytes.NewBuffer([]byte(query))

	req, err := http.NewRequest("POST", "https://api.linear.app/graphql", body)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", key)

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	var response Response
	jsonErr := json.NewDecoder(res.Body).Decode(&response)

	if jsonErr != nil {
		return nil, jsonErr
	}

	return response.Data.Issues.Nodes, nil

}
