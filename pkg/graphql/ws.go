// Copyright © 2020 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"emperror.dev/errors"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
)

const (
	connectionInitMsg = "connection_init" // Client -> Server
	startMsg          = "start"           // Client -> Server
	connectionAckMsg  = "connection_ack"  // Server -> Client
	connectionKaMsg   = "ka"              // Server -> Client
	dataMsg           = "data"            // Server -> Client
	errorMsg          = "error"           // Server -> Client
)

type operationMessage struct {
	Payload json.RawMessage `json:"payload,omitempty"`
	ID      string          `json:"id,omitempty"`
	Type    string          `json:"type"`
}

type WSClient struct {
	endpoint string

	httpClient *http.Client
}

type ClientOption func(*WSClient)

func WithHTTPClient(httpclient *http.Client) ClientOption {
	return func(client *WSClient) {
		client.httpClient = httpclient
	}
}

func NewWSClient(endpoint string, opts ...ClientOption) *WSClient {
	return &WSClient{
		endpoint: endpoint,
	}
}

func (c *WSClient) Subscribe(ctx context.Context, req *Request, resp chan interface{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return c.runWithJSON(ctx, req, resp)
}

func (c *WSClient) runWithJSON(ctx context.Context, req *Request, resp chan interface{}) error {
	var requestBody bytes.Buffer
	requestBodyObj := struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}{
		Query:     req.query,
		Variables: req.vars,
	}
	if err := json.NewEncoder(&requestBody).Encode(requestBodyObj); err != nil {
		return errors.Wrap(err, "encode body")
	}

	u, err := url.Parse(c.endpoint)
	if err != nil {
		return errors.Wrap(err, "could not parse endpoint URL")
	}
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	}

	dialer := websocket.DefaultDialer
	if c.httpClient != nil {
		if t, ok := c.httpClient.Transport.(*http.Transport); ok && t.TLSClientConfig != nil {
			dialer.TLSClientConfig = t.TLSClientConfig
		}
	}

	wsc, r, err := dialer.Dial(u.String(), req.GetHeader().Clone())
	if err != nil {
		return errors.WrapIf(err, "could not connect to websocket")
	}
	defer r.Body.Close()
	defer wsc.Close()

	initMessage := operationMessage{Type: connectionInitMsg}

	err = wsc.WriteJSON(initMessage)
	if err != nil {
		return errors.WrapIf(err, "could not write message to websocket")
	}

	err = wsc.WriteJSON(operationMessage{Type: startMsg, ID: "1", Payload: requestBody.Bytes()})
	if err != nil {
		return errors.WrapIf(err, "could not write message to websocket")
	}

	type BaseMessage struct {
		ID      string
		Payload interface{}
		Type    string
	}

	for {
		// var a BaseMessage
		t, r, err := wsc.NextReader()
		if err != nil {
			return errors.WrapIf(err, "could not read next data from socket")
		}
		if t != websocket.TextMessage {
			continue
		}

		message, err := ioutil.ReadAll(r)
		if err != nil {
			return errors.WrapIf(err, "could not read all data from socket")
		}

		var msg BaseMessage
		err = json.Unmarshal(message, &msg)
		if err != nil {
			return errors.WrapIf(err, "could not unmarshal data")
		}

		type graphErr struct {
			Message string
		}
		type Errors []graphErr

		switch msg.Type {
		case connectionAckMsg, connectionKaMsg:
		case "complete":
			return nil
		case errorMsg:
			var errs Errors
			err = mapstructure.Decode(msg.Payload, &errs)
			if err != nil {
				return err
			}
			if len(errs) > 0 {
				return errors.New(errs[0].Message)
			}
		case dataMsg:
			select {
			default:
				if data, ok := msg.Payload.(map[string]interface{}); ok {
					if data[dataMsg] != nil {
						resp <- data[dataMsg]
					}
					if data["errors"] != nil {
						var errs Errors
						err = mapstructure.Decode(data["errors"], &errs)
						if err != nil {
							return err
						}
						if len(errs) > 0 {
							return errors.New(errs[0].Message)
						}
					}
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}
