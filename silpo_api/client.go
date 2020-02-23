package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/uuid"
)

// Token describes json from with tokens
type Token struct {
	Expireat int    `json:"expireat"`
	Value    string `json:"value"`
}

// Tokens describes json from with tokens
type Tokens struct {
	AccessToken  Token `json:"accessToken"`
	RefreshToken Token `json:"refreshToken"`
}

// Client contains method to work with API
type Client struct {
	Tokens Tokens
}

type rpcPayload struct {
	Data   interface{} `json:"Data`
	Method method      `json:"Method`
}

type tokenResponse struct {
	Tokens Tokens `json:"tokens"`
}

type sendOPTData struct {
	ForceUpdate bool   `json:"forceUpdate"`
	GUID        string `json:"guid"`
	Phone       string `json:"phone"`
	WithPoints  bool   `json:"withPoints"`
}

type comfirmOPTData struct {
	ForceUpdate bool   `json:"forceUpdate"`
	GUID        string `json:"guid"`
	OTPCode     string `json:"otpCode"`
	Phone       string `json:"phone"`
	WithPoints  bool   `json:"withPoints"`
}

type defautlData struct {
	ForceUpdate bool `json:"forceUpdate"`
	WithPoints  bool `json:"withPoints"`
}

const (
	rpcURL = "https://api.sm.silpo.ua/api/2.0/exec/FZGlobal/"
)

type method string

const (
	sendOTP              method = "SendOTP"
	confirmationOtpV2    method = "ConfirmationOtp_V2"
	getLastChequeHeaders method = "GetLastChequeHeaders"
	getChequesInfos      method = "GetChequesInfos"
)

func (c *Client) requestBuilder(payload interface{}, m method) (*http.Request, error) {
	rpcData := rpcPayload{
		payload,
		m,
	}
	rpcDataSerialised, err := json.Marshal(rpcData)
	if err != nil {
		return nil, err
	}

	body := bytes.NewReader(rpcDataSerialised)

	req, err := http.NewRequest("POST", rpcURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *Client) exec(req *http.Request) ([]byte, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return respBody, nil
}

func (c *Client) performCall(payload interface{}, m method) ([]byte, error) {
	req, err := c.requestBuilder(payload, m)
	if err != nil {
		return []byte{}, err
	}
	return c.exec(req)

}

func (c *Client) performCallWithAuth(payload interface{}, m method) ([]byte, error) {
	req, err := c.requestBuilder(payload, m)
	if err != nil {
		return []byte{}, err
	}
	// TODO: check if token is valid, refresh token
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.Tokens.AccessToken.Value))
	return c.exec(req)
}

// SendOTP starts sing in process sending sms to phone numbers
func (c *Client) SendOTP(guid uuid.UUID, phone string) ([]byte, error) {
	payload := sendOPTData{false, guid.String(), phone, false}
	return c.performCall(payload, sendOTP)
}

// ConfirmationOtp finish sing in process getting tokens
func (c *Client) ConfirmationOtp(guid uuid.UUID, phone string, otpCode string) ([]byte, error) {
	payload := comfirmOPTData{false, guid.String(), otpCode, phone, false}
	return c.performCall(payload, confirmationOtpV2)
}

// GetLastChequeHeaders get list of receipts
func (c *Client) GetLastChequeHeaders() ([]byte, error) {
	payload := defautlData{false, false}
	return c.performCallWithAuth(payload, getLastChequeHeaders)
}

// GetChequesInfos get details of receipt
func (c *Client) GetChequesInfos() ([]byte, error) {
	// {
	// 	"Data": {
	// 		"forceUpdate": false,
	// 		"identities": [{
	// 			"chequeId": 9994795,
	// 			"created": "2020-02-21T18:23:07",
	// 			"filId": 2042
	// 		}],
	// 		"withPoints": false
	// 	},
	// 	"Method": "GetChequesInfos"
	// }
	payload := defautlData{false, false}
	return c.performCallWithAuth(payload, getChequesInfos)
}

func (c *Client) setTokens(data []byte) error {
	tokensResponse := tokenResponse{}
	err := json.Unmarshal(data, &tokensResponse)
	if err != nil {
		return err
	}
	c.Tokens = tokensResponse.Tokens
	return nil
}
