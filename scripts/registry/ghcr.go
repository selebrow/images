package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type tokenResponse struct {
	Token string `json:"token"`
}

func GenerateGHCRToken(project, image string) (string, error) {
	url := fmt.Sprintf(
		"https://ghcr.io/token?service=ghcr.io&scope=repository:%s/%s:pull",
		project,
		image,
	)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tokenResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", nil
	}

	return tokenResp.Token, nil
}
