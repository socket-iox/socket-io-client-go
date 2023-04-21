package transport

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

type Polling struct {
	client    *http.Client
	url       string
	nameSpace *string
	sid       string
	isClosed  bool
	params    *TransportParams

	recv chan []string
	send chan []string
}

func NewPolling(url string, nameSpace *string) Polling {
	client := &http.Client{}

	return Polling{
		client:    client,
		url:       url + "?EIO=4&transport=polling",
		nameSpace: nameSpace,
		isClosed:  false,
		recv:      make(chan []string, ChannelBufferSize),
		send:      make(chan []string, ChannelBufferSize)}
}

func (p *Polling) Open() error {
	p.client = &http.Client{}
	resp, err := p.client.Get(p.url)
	if err != nil {
		return err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()

	body := string(bodyBytes)

	if string(body[0]) != OpenPacket {
		return errors.New("should be open packet")
	}

	// decode params
	var params TransportParams
	if err := json.Unmarshal([]byte(body[1:]), &params); err != nil {
		return err
	}
	p.params = &params

	p.url += "&sid=" + params.Sid

	// open namespace
	var openMsg string
	if p.nameSpace != nil {
		openMsg = OpenNamespace + *p.nameSpace + ","
	} else {
		openMsg = OpenNamespace
	}
	resp, err = p.client.Post(p.url, "application/text", strings.NewReader(openMsg))
	if err != nil {
		return err
	}

	p.recv <- []string{"connect"}

	return nil
}

func (p *Polling) Read() (message []string, err error) {
	select {
	case message = <-p.recv:
		return message, nil
	default:
	}

	resp, err := p.client.Get(p.url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("read request failed")
	}

	bodyBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}
	_ = resp.Body.Close()

	payloads := bytes.Split(bodyBytes, []byte("\x1e"))

	for _, payload := range payloads {
		payloadStr := string(payload)
		if strings.HasPrefix(payloadStr, NewMessage) {
			payloadStr = payloadStr[2:]
			if strings.Contains(payloadStr, *p.nameSpace) {
				payloadStr = strings.Split(payloadStr, *p.nameSpace+",")[1]
			}
			var message []string
			if err := json.Unmarshal([]byte(payloadStr), &message); err != nil {
				return nil, err
			}
			p.recv <- message
		}
		if payloadStr == PingPacket {
			_ = p.Write(PongPacket)
		}
	}

	select {
	case message = <-p.recv:
		return message, nil
	default:
		return nil, nil
	}
}

func (p *Polling) Write(message string) error {
	resp, err := p.client.Post(p.url, "application/text", strings.NewReader(message))
	if err != nil {
		return err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()

	bodyStr := string(bodyBytes)
	if bodyStr != "ok" {
		return errors.New("emit failed: " + bodyStr)
	}

	return nil
}
func (p *Polling) Close() {
	p.recv <- []string{"disconnect"}
	p.isClosed = true
}
