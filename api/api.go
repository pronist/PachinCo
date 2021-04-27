package api

import "github.com/pronist/upbit/gateway"

type API struct {
	Client *gateway.Client
	QuotationClient *gateway.QuotationClient
}
