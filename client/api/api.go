package api

import "github.com/pronist/upbit/client"

type API struct {
	Client          *client.Client
	QuotationClient *client.QuotationClient
}
