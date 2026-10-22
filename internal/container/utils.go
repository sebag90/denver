package container

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Container struct {
	ID     string   `json:"Id"`
	Names  []string `json:"Names"`
	Image  string   `json:"Image"`
	State  string   `json:"State"`
	Status string   `json:"Status"`
}

const (
	apiVersion   = "1.54"
	baseApiUrl   = "http://container/v" + apiVersion
	defaultImage = "ghcr.io/sebag90/devenv:latest"
)

type PullProgress struct {
	Status         string `json:"status"`
	ID             string `json:"id"`
	Progress       string `json:"progress"`
	ProgressDetail struct {
		Current int64 `json:"current"`
		Total   int64 `json:"total"`
	} `json:"progressDetail"`
}

func getRunningContainer(client *http.Client) string {
	req, err := http.NewRequest(
		http.MethodGet,
		baseApiUrl+"/containers/json",
		nil,
	)
	if err != nil {
		panic(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("Podman returned %s", resp.Status))
	}

	var containers []Container
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		panic(err)
	}

	return containers[0].ID
}

func ListContainers(client *http.Client, name string) string {
	req, err := http.NewRequest(
		http.MethodGet,
		baseApiUrl+"/containers/json",
		nil,
	)
	if err != nil {
		panic(err)
	}

	q := req.URL.Query()
	filters, err := json.Marshal(map[string][]string{"name": {name}})
	if err != nil {
		panic(err)
	}
	q.Set("filters", string(filters))
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		panic(fmt.Sprintf("Podman returned %s: %s", resp.Status, body))
	}

	var containers []Container
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		panic(err)
	}

	for _, c := range containers {
		fmt.Printf("%s %v %s %s\n",
			c.ID[:12],
			c.Names,
			c.Image,
			c.State,
		)
	}
	return containers[0].ID
}

func hash(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	return strconv.FormatUint(uint64(h.Sum32()), 10)
}

func NameContainer() string {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	return fmt.Sprintf("denver.%s", hash(path))
}

func downloadImage(client *http.Client, image string) {
	req, err := http.NewRequest(
		http.MethodPost,
		baseApiUrl+"/images/create",
		nil,
	)
	if err != nil {
		panic(err)
	}

	q := req.URL.Query()
	q.Set("fromImage", image)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Docker error: %s: %s", resp.Status, body)
	}

	decoder := json.NewDecoder(resp.Body)

	for {
		var progress PullProgress

		err := decoder.Decode(&progress)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			panic(err)
		}

		fmt.Printf("%-20s %-30s %s\n",
			progress.ID,
			progress.Status,
			progress.Progress,
		)
	}
}
