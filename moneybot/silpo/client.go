package silpo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/uuid"
)

// Client contains method to work with API
type Client struct {
	Tokens Tokens
}

// Token describes json from with tokens
type Token struct {
	Value string `json:"value"`
}

// Tokens describes json from with tokens
type Tokens struct {
	AccessToken  Token `json:"accessToken"`
	RefreshToken Token `json:"refreshToken"`
}

// Identities stores ID of receipt
type Identities struct {
	ChequeID int    `json:"chequeId"`
	Created  string `json:"created"`
	FilID    int    `json:"filId"`
}

type rpcPayload struct {
	Data   interface{} `json:"Data"`
	Method method      `json:"Method"`
}

type tokenResponse struct {
	Tokens Tokens `json:"tokens"`
}

type sendOPTDataRequst struct {
	ForceUpdate bool   `json:"forceUpdate"`
	GUID        string `json:"guid"`
	Phone       string `json:"phone"`
	WithPoints  bool   `json:"withPoints"`
}

type comfirmOPTDataRequest struct {
	ForceUpdate bool   `json:"forceUpdate"`
	GUID        string `json:"guid"`
	OTPCode     string `json:"otpCode"`
	Phone       string `json:"phone"`
	WithPoints  bool   `json:"withPoints"`
}

type defaultDataRequest struct {
	ForceUpdate bool `json:"forceUpdate"`
	WithPoints  bool `json:"withPoints"`
}

type detailReceiptDataRequest struct {
	ForceUpdate bool         `json:"forceUpdate"`
	Identities  []Identities `json:"identities"`
	WithPoints  bool         `json:"withPoints"`
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
	refreshToken         method = "RefreshToken"
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
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.Tokens.AccessToken.Value))
	return c.exec(req)
}

// SendOTP starts sing in process sending sms to phone numbers
func (c *Client) SendOTP(guid uuid.UUID, phone string) ([]byte, error) {
	payload := sendOPTDataRequst{false, guid.String(), phone, false}
	return c.performCall(payload, sendOTP)
}

// ConfirmationOtp finish sing in process getting tokens and sets to client
func (c *Client) ConfirmationOtp(guid uuid.UUID, phone string, otpCode string) ([]byte, error) {
	payload := comfirmOPTDataRequest{false, guid.String(), otpCode, phone, false}
	tokenResponse, err := c.performCall(payload, confirmationOtpV2)
	if err != nil {
		return []byte{}, err
	}

	err = c.SetTokens(tokenResponse)
	if err != nil {
		return []byte{}, err
	}

	return tokenResponse, nil
}

// RefreshToken gets new tokens instead of expired
func (c *Client) RefreshToken() ([]byte, error) {
	payload := defaultDataRequest{false, false}
	req, err := c.requestBuilder(payload, refreshToken)
	if err != nil {
		return []byte{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.Tokens.RefreshToken.Value))
	return c.exec(req)
}

// GetLastChequeHeaders get list of receipts
func (c *Client) GetLastChequeHeaders() ([]byte, error) {
	payload := defaultDataRequest{false, false}
	return c.performCallWithAuth(payload, getLastChequeHeaders)
}

// GetChequesInfos get details of receipt
func (c *Client) GetChequesInfos(chequeID int, created string, filID int) ([]byte, error) {
	payload := detailReceiptDataRequest{false, []Identities{{chequeID, created, filID}}, false}
	return c.performCallWithAuth(payload, getChequesInfos)
}

// ParseToken gets tokens from response
func (c *Client) ParseToken(data []byte) (Tokens, error) {
	tokensResponse := tokenResponse{}
	err := json.Unmarshal(data, &tokensResponse)
	if err != nil {
		return tokensResponse.Tokens, err
	}
	return tokensResponse.Tokens, nil
}

// SetTokens sets tokens to client
func (c *Client) SetTokens(data []byte) error {
	tokens, err := c.ParseToken(data)
	if err != nil {
		return err
	}
	c.Tokens = tokens
	return nil
}
