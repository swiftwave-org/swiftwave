package udp_proxy_manager

import (
	"encoding/json"
	"errors"
	"io"
	"log"
)

func (m Manager) List() ([]Proxy, error) {
	req, err := m.getRequest("/proxy/list")
	if err != nil {
		return nil, errors.New("failed to send request")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("failed to close body in list udpproxy function")
		}
	}(req.Body)
	if req.StatusCode != 200 {
		return nil, errors.New("failed to get proxy list")
	}
	var response []Proxy
	err = json.NewDecoder(req.Body).Decode(&response)
	if err != nil {
		return nil, errors.New("failed to decode response")
	}
	return response, nil
}
