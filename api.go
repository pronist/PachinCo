package upbit

import (
	"github.com/pronist/upbit/client"
	"github.com/pronist/upbit/client/api"
	"net/http"
)

var API *api.API

func init() {
	API = &api.API{
		Client:          &client.Client{AccessKey: Config.Keypair.AccessKey, SecretKey: Config.Keypair.SecretKey},
		QuotationClient: &client.QuotationClient{Client: &http.Client{}},
	}
}
