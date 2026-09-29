package container

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func CreateContainer(client *http.Client, imageName string) {
	payload := map[string]any{
		"Image":      imageName,
		"Env":        commonEnv,
		"Tty":        true,
		"OpenStdin":  true,
		"HostConfig": hostConfig,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		baseApiUrl+"/containers/create",
		bytes.NewReader(jsonData),
	)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")

	q := req.URL.Query()
	q.Set("name", "ciao")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Podman create error: %s: %s", resp.Status, body)
	}
}
