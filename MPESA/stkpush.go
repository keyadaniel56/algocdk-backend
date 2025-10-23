package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"strings"
	"time"
)

// ======================
// CONFIGURATION
// ======================
const (
	ConsumerKey     = "6sd9d5Nr3cwwOsB7TGbGD7ssPswSboOInAezGSaxTGAIe76O"
	ConsumerSecret  = "FxCb7SwfDXlKAyOOuvmtctDfF4IXwpNTuAXC5NUUUxzV2Sdg06YED5zyOVK1E8qd"
	ShortCode       = "174379"
	PassKey         = "bfb279f9aa9bdbcf158e97dd71a467cd2e0c893059b10f78e6b72ada1ed2c919"
	AccountRef      = "TestAccount"
	TransactionDesc = "Daraja Go SDK STK Push"
	SandboxBaseURL  = "https://sandbox.safaricom.co.ke"
)

// ======================
// STRUCTS
// ======================
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   string `json:"expires_in"`
}

type STKPushRequest struct {
	BusinessShortCode string `json:"BusinessShortCode"`
	Password          string `json:"Password"`
	Timestamp         string `json:"Timestamp"`
	TransactionType   string `json:"TransactionType"`
	Amount            int    `json:"Amount"`
	PartyA            string `json:"PartyA"`
	PartyB            string `json:"PartyB"`
	PhoneNumber       string `json:"PhoneNumber"`
	CallBackURL       string `json:"CallBackURL"`
	AccountReference  string `json:"AccountReference"`
	TransactionDesc   string `json:"TransactionDesc"`
}

type STKPushResponse struct {
	MerchantRequestID   string `json:"MerchantRequestID"`
	CheckoutRequestID   string `json:"CheckoutRequestID"`
	ResponseCode        string `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
	CustomerMessage     string `json:"CustomerMessage"`
}

type STKQueryResponse struct {
	ResponseCode        string `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
	ResultCode          string `json:"ResultCode"`
	ResultDesc          string `json:"ResultDesc"`
	CheckoutRequestID   string `json:"CheckoutRequestID"`
	MerchantRequestID   string `json:"MerchantRequestID"`
}

// ======================
// HELPERS
// ======================
func formatPhoneNumber(number string) (string, error) {
	number = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(number, "+", ""), " ", ""))
	if strings.HasPrefix(number, "0") {
		return "254" + number[1:], nil
	}
	if strings.HasPrefix(number, "254") {
		return number, nil
	}
	return "", fmt.Errorf("invalid phone number format")
}

// ======================
// ACCESS TOKEN
// ======================
func getAccessToken() (string, error) {
	url := SandboxBaseURL + "/oauth/v1/generate?grant_type=client_credentials"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	auth := base64.StdEncoding.EncodeToString([]byte(ConsumerKey + ":" + ConsumerSecret))
	req.Header.Add("Authorization", "Basic "+auth)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	var token TokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return "", err
	}
	return token.AccessToken, nil
}

// ======================
// STK PUSH REQUEST
// ======================
func SendSTKPush(phone string, amount int) (*STKPushResponse, error) {
	formatted, err := formatPhoneNumber(phone)
	if err != nil {
		return nil, err
	}

	token, err := getAccessToken()
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().Format("20060102150405")
	password := base64.StdEncoding.EncodeToString([]byte(ShortCode + PassKey + timestamp))

	payload := STKPushRequest{
		BusinessShortCode: ShortCode,
		Password:          password,
		Timestamp:         timestamp,
		TransactionType:   "CustomerPayBillOnline",
		Amount:            amount,
		PartyA:            formatted,
		PartyB:            ShortCode,
		PhoneNumber:       formatted,
		CallBackURL:       "https://example.com/fakecallback", // Fake URL, ignored in sandbox
		AccountReference:  AccountRef,
		TransactionDesc:   TransactionDesc,
	}

	jsonData, _ := json.Marshal(payload)
	url := SandboxBaseURL + "/mpesa/stkpush/v1/processrequest"

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	var response STKPushResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// ======================
// STK QUERY
// ======================
func QuerySTKPush(checkoutID string) (*STKQueryResponse, error) {
	token, err := getAccessToken()
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().Format("20060102150405")
	password := base64.StdEncoding.EncodeToString([]byte(ShortCode + PassKey + timestamp))

	data := map[string]string{
		"BusinessShortCode": ShortCode,
		"Password":          password,
		"Timestamp":         timestamp,
		"CheckoutRequestID": checkoutID,
	}

	jsonData, _ := json.Marshal(data)
	url := SandboxBaseURL + "/mpesa/stkpushquery/v1/query"

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	var response STKQueryResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// ======================
// UTILITY
// ======================
func PrintJSON(v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(data))
}

// ======================
// MAIN
// ======================
func main() {
	phone := "254708374149" // Sandbox test number
	amount := 10            // Test amount

	fmt.Println("Sending STK Push...")
	resp, err := SendSTKPush(phone, amount)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("STK Push Response:")
	PrintJSON(resp)

	// Query STK status after 5 seconds
	if resp.CheckoutRequestID != "" {
		time.Sleep(5 * time.Second)
		fmt.Println("\nQuerying STK Push status...")
		status, err := QuerySTKPush(resp.CheckoutRequestID)
		if err != nil {
			fmt.Println("Error querying STK:", err)
			return
		}
		PrintJSON(status)
	}
}
