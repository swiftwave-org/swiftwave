package udp_proxy_manager

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
)

func (m Manager) Exist(proxy Proxy) (bool, error) {
	jsonMarshal, err := json.Marshal(proxy)
	if err != nil {
		return false, errors.New("invalid payload")
	}

	req, err := m.postRequest("/proxy/exists", bytes.NewReader(jsonMarshal))
	if err != nil {
		return false, errors.New("failed to send request")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("failed to close body in exist udpproxy function")
		}
	}(req.Body)
	if req.StatusCode != 200 {
		return false, errors.New("failed to check proxy existence")
	}
	var response ExistProxyResponse
	err = json.NewDecoder(req.Body).Decode(&response)
	if err != nil {
		return false, errors.New("failed to decode response")
	}
	return response.Exist, nil
}
