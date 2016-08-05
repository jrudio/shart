package main

import (
	"net/http"
	"time"
)

func get(query string) (*http.Response, error) {
	client := http.Client{
		Timeout: 3 * time.Second,
	}

	req, err := http.NewRequest("GET", query, nil)

	if err != nil {
		return &http.Response{}, err
	}

	return client.Do(req)
}
