package container

import (
	"fmt"
	"net/http"
)

func StartContainer(client *http.Client, containerId string) {
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/containers/%s/start", baseApiUrl, containerId),
		nil,
	)

	_, err = client.Do(req)

	if err != nil {
		panic(err)
	}
}
