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
	ConsumerKey    = "6sd9d5Nr3cwwOsB7TGbGD7ssPswSboOInAezGSaxTGAIe76O"
	ConsumerSecret = "FxCb7SwfDXlKAyOOuvmtctDfF4IXwpNTuAXC5NUUUxzV2Sdg06YED5zyOVK1E8qd"
	ShortCode      = "174379"
	PassKey        = "bfb279f9aa9bdbcf158e97dd71a467cd2e0c893059b10f78e6b72ada1ed2c919"
	AccountRef     = "AlgoCdk"
	SandboxBaseURL = "https://sandbox.safaricom.co.ke"
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

// Payment struct
type Payment struct {
	Type       string // "purchase" or "subscription"
	Item       string
	Amount     int
	Frequency  string // "week", "month" for subscriptions
	BuyerName  string
	BuyerPhone string
	BotID      string
}

// Bot struct
type Bot struct {
	BotID     string
	OwnerName string
	IsActive  bool
	StartDate time.Time
	EndDate   time.Time
}

// In-memory bot storage (for example)
var bots = map[string]*Bot{}

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
// STK PUSH
// ======================
func SendSTKPushWithDesc(phone string, amount int, desc string) (*STKPushResponse, error) {
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

	if len(desc) > 50 {
		desc = desc[:50]
	}

	payload := STKPushRequest{
		BusinessShortCode: ShortCode,
		Password:          password,
		Timestamp:         timestamp,
		TransactionType:   "CustomerPayBillOnline",
		Amount:            amount,
		PartyA:            formatted,
		PartyB:            ShortCode,
		PhoneNumber:       formatted,
		CallBackURL:       "https://example.com/fakecallback",
		AccountReference:  AccountRef,
		TransactionDesc:   desc,
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
		fmt.Println("Raw response:", string(body))
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
// PAYMENT PROCESSOR
// ======================
func ProcessPayment(p Payment) (*STKPushResponse, error) {
	desc := ""
	if p.Type == "purchase" {
		desc = fmt.Sprintf("%s buying %s", p.BuyerName, p.Item)
	} else if p.Type == "subscription" {
		desc = fmt.Sprintf("%s subscribing to %s (%s)", p.BuyerName, p.Item, p.Frequency)
	} else {
		desc = fmt.Sprintf("%s paying for %s", p.BuyerName, p.Item)
	}

	return SendSTKPushWithDesc(p.BuyerPhone, p.Amount, desc)
}

// Activate bot after successful payment
func ActivateBot(p Payment) {
	now := time.Now()
	bot := &Bot{
		BotID:     p.BotID,
		OwnerName: p.BuyerName,
		IsActive:  true,
		StartDate: now,
	}

	if p.Type == "subscription" {
		// Set end date depending on frequency
		if p.Frequency == "week" {
			bot.EndDate = now.AddDate(0, 0, 7)
		} else if p.Frequency == "month" {
			bot.EndDate = now.AddDate(0, 1, 0)
		}
	}

	bots[p.BotID] = bot
	fmt.Printf("Bot %s activated for %s\n", bot.BotID, bot.OwnerName)
}

// Check subscription expirations
func CheckSubscriptions() {
	now := time.Now()
	for _, bot := range bots {
		if bot.IsActive && !bot.EndDate.IsZero() && bot.EndDate.Before(now) {
			bot.IsActive = false
			fmt.Printf("Subscription expired for bot %s, locking it again\n", bot.BotID)
		}
	}
}

// Utility
func PrintJSON(v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(data))
}

// ======================
// MAIN
// ======================
func main() {
	// Example: purchase
	purchase := Payment{
		Type:       "purchase",
		Item:       "Bot X",
		Amount:     1000,
		BuyerName:  "Alice",
		BuyerPhone: "254798062459",
		BotID:      "BOT123",
	}

	// Example: subscription
	subscription := Payment{
		Type:       "subscription",
		Item:       "Bot Y",
		Amount:     200,
		Frequency:  "week",
		BuyerName:  "Bob",
		BuyerPhone: "254712345678",
		BotID:      "BOT456",
	}

	// Process purchase
	fmt.Println("Processing purchase...")
	resp, err := ProcessPayment(purchase)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	PrintJSON(resp)

	if resp.CheckoutRequestID != "" {
		for i := 0; i < 5; i++ {
			time.Sleep(5 * time.Second)
			fmt.Printf("\nQuery attempt %d...\n", i+1)
			status, _ := QuerySTKPush(resp.CheckoutRequestID)
			PrintJSON(status)
			if status.ResultCode == "0" {
				ActivateBot(purchase)
				break
			}
		}
	}

	// Process subscription
	fmt.Println("\nProcessing subscription...")
	respSub, err := ProcessPayment(subscription)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	PrintJSON(respSub)

	if respSub.CheckoutRequestID != "" {
		for i := 0; i < 5; i++ {
			time.Sleep(5 * time.Second)
			fmt.Printf("\nQuery attempt %d...\n", i+1)
			status, _ := QuerySTKPush(respSub.CheckoutRequestID)
			PrintJSON(status)
			if status.ResultCode == "0" {
				ActivateBot(subscription)
				break
			}
		}
	}

	// Simulate checking subscription expirations
	fmt.Println("\nChecking subscription expirations...")
	CheckSubscriptions()
}
