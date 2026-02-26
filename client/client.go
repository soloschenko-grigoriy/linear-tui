package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Issue struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	State       struct {
		Name     string  `json:"name"`
		Position float64 `json:"position"`
	} `json:"state"`
	Priority int `json:"priority"`
}

type graphqlReqBody struct {
	Query string `json:"query"`
}

type Response struct {
	Data struct {
		Issues struct {
			Nodes []Issue `json:"nodes"`
		} `json:"issues"`
	} `json:"data"`
}

func FetchIssues(includeDone bool) ([]Issue, error) {
	key := os.Getenv("LINEAR_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("LINEAR_API_KEY not set")
	}

	var states = `["In Progress", "In Review", "Todo", "Done", "Pending"]`
	if !includeDone {
		states = `["In Progress", "In Review", "Todo", "Pending"]`
	}

	query := fmt.Sprintf(`{ issues(filter: { assignee: { isMe: { eq: true } }, cycle: { isActive: { eq: true } }, state: { name: { in: %s } } }, first: 100) { nodes { id title description state { name position } priority } } }`, states)

	reqBody := graphqlReqBody{
		Query: query,
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	body := bytes.NewBuffer(jsonBytes)

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
