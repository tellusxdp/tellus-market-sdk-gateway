package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type CountRequest struct {
	ProductID string `json:"product_id"`
	UserID    string `json:"user_id"`
	Token     string `json:"token"`
	RequestID string `json:"request_id"`
}

func (s *Server) StartCountRequestLoop() chan<- CountRequest {
	c := make(chan CountRequest, 100)

	go func(c <-chan CountRequest) {
		for {
			r := <-c
			s.Logger.Debugf("Request count: %s", r.RequestID)
			err := s.count(r)
			if err != nil {
				s.Logger.Errorf("Count request error: %s", err.Error())
			}
		}
	}(c)

	return c
}

func (s *Server) count(r CountRequest) error {
	body, _ := json.Marshal(&r)

	req, err := http.NewRequest("POST", s.Config.CounterURL, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.Config.APIKey)

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
