package udp_proxy_manager

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
)

func (m Manager) Add(proxy Proxy, restrictedPorts []int) error {
	if IsPortRestrictedForManualConfig(proxy.Port, restrictedPorts) {
		return errors.New("port is restricted")
	}
	jsonMarshal, err := json.Marshal(proxy)
	if err != nil {
		return errors.New("invalid payload")
	}

	req, err := m.postRequest("/proxy/add", bytes.NewReader(jsonMarshal))
	if err != nil {
		return errors.New("failed to send request")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("failed to close body in add udpproxy function")
		}
	}(req.Body)
	if req.StatusCode != 200 {
		return errors.New("failed to add proxy")
	}
	var response AddProxyResponse
	err = json.NewDecoder(req.Body).Decode(&response)
	if err != nil {
		return errors.New("failed to decode response")
	}
	if !response.Success {
		return errors.New(response.Error)
	}
	return nil
}
