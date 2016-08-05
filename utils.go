package main

import (
	"bytes"
	"net/http"
	"net/url"
	"time"
)

// utils.go holds network utils and function helpers

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

func post(query string, body []byte) (*http.Response, error) {
	client := http.Client{
		Timeout: 3 * time.Second,
	}

	req, err := http.NewRequest("POST", query, bytes.NewBuffer(body))

	if err != nil {
		return &http.Response{}, err
	}

	req.Header.Set("Content-type", "application/json")

	return client.Do(req)
}

func encodeURL(str string) (string, error) {
	u, err := url.Parse(str)

	if err != nil {
		return "", err
	}

	return u.String(), nil
}
