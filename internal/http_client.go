package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	p2phttp "github.com/libp2p/go-libp2p-http"
	"github.com/libp2p/go-libp2p/core/host"
	"ipfs-sharing/gen/model"
	"log"
	"net/http"
)

type HttpClient struct {
	Cl *http.Client
}

func NewHttpClient(host host.Host) *HttpClient {
	tr := &http.Transport{}
	tr.RegisterProtocol("libp2p", p2phttp.NewTransport(host))
	cl := &http.Client{Transport: tr}

	return &HttpClient{Cl: cl}
}

func (hc *HttpClient) PostJson(url string, structData interface{}) (*http.Response, error) {
	postBody, err := json.Marshal(structData)
	if err != nil {
		return nil, err
	}
	resp, err := hc.Cl.Post("libp2p://"+url, "application/json", bytes.NewBuffer(postBody))
	if err != nil {
		return nil, err
	}

	return resp, err
}

func (hc *HttpClient) GetChildren(address string, id int32, isRecursive bool) (contents []model.Contents, err error) {

	r, err := hc.PostJson(fmt.Sprint(address, "/content/children?id=", id, "&recursive=", fmt.Sprintf("%t", isRecursive)), nil)
	if err != nil {
		return
	}

	err = json.NewDecoder(r.Body).Decode(&contents)
	if err != nil {
		log.Println(err)
		return
	}

	for i, _ := range contents {
		contents[i].From = address
		contents[i].Status = ""
	}

	return
}
