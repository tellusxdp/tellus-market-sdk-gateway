package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type CountRequest struct {
	ToolID    string `json:"tool_id"`
	UserID    string `json:"user_id"`
	Token     string `json:"token"`
	RequestID string `json:"request_id"`
}

func (s *Server) Count(r CountRequest) error {
	body, _ := json.Marshal(&r)

	req, err := http.NewRequest("POST", s.CounterURL, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.APIKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		body, _ := ioutil.ReadAll(res.Body)
		return fmt.Errorf("Invalid Status Code %d %s", res.StatusCode, string(body))
	}

	return nil
}
