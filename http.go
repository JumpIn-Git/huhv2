package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var Client = &http.Client{Timeout: 20 * time.Second}

func HttpReq(req *http.Request) (*http.Response, error) {
	resp, err := Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Request failed: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return resp, fmt.Errorf("Unexpected status code: %d", resp.StatusCode)
	}
	return resp, nil
}

func HttpJson(req *http.Request, v any) error {
	resp, err := HttpReq(req)
	if err != nil {
		if resp != nil {
			resp.Body.Close()
		}
		return err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("Malformed response: %w", err)
	}
	return nil
}
