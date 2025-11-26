package components

import (
	"bytes"
	"crypto/tls"
	"io"
	"net/http"
)

var clientHTTP = http.DefaultClient

var clientHTTPS = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

func get_client(useSSL bool) *http.Client {
	if useSSL {
		return clientHTTPS
	}
	return clientHTTP
}

func build_url(host string, path string, useSSL bool) string {
	var scheme string = "http://"

	if useSSL {
		scheme = "https://"
	}

	return scheme + host + path
}

func do_get(url string, useSSL bool) ([]byte, error) {
	client := get_client(useSSL)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func do_head_get(url string, headers map[string]string, useSSL bool) ([]byte, error) {
	client := get_client(useSSL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func do_post(url string, data []byte, useSSL bool) ([]byte, error) {
	client := get_client(useSSL)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func do_head_post(url string, body []byte, headers map[string]string, useSSL bool) ([]byte, error) {
	client := get_client(useSSL)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
