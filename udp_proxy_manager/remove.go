package udp_proxy_manager

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
)

func (m Manager) Remove(proxy Proxy) error {
	jsonMarshal, err := json.Marshal(proxy)
	if err != nil {
		return errors.New("invalid payload")
	}

	req, err := m.postRequest("/proxy/remove", bytes.NewReader(jsonMarshal))
	if err != nil {
		return errors.New("failed to send request")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("failed to close body in remove udpproxy function")
		}
	}(req.Body)
	if req.StatusCode != 200 {
		return errors.New("failed to remove proxy")
	}
	var response RemoveProxyResponse
	err = json.NewDecoder(req.Body).Decode(&response)
	if err != nil {
		return errors.New("failed to decode response")
	}
	if !response.Success {
		return errors.New(response.Error)
	}
	return nil
}
