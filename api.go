package upbit

import (
	"github.com/pronist/upbit/client"
	"github.com/pronist/upbit/client/api"
	"net/http"
)

const (
	accessKey = "UyGWYAEVN3PRDDo3Y3pJnV6DWn69k17gVs1X47p4"
	secretKey = "2FjMz4yBOuHqzpwGUdkEu0WJF5g30Z8Wx71cJbxn"
)

var API = &api.API{
	Client:          &client.Client{Client: &http.Client{}, AccessKey: accessKey, SecretKey: secretKey},
	QuotationClient: &client.QuotationClient{Client: &http.Client{}},
}
