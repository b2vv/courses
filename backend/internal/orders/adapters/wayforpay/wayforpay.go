package wayforpay

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"courses/internal/orders/domain"
	"courses/internal/orders/ports"
)

const (
	ProviderName = "wayforpay"
	PayURL       = "https://secure.wayforpay.com/pay"
	StatusURL    = "https://api.wayforpay.com/api"
	maxPayload   = 64 * 1024
)

type Config struct {
	MerchantAccount string
	SecretKey       string
	MerchantDomain  string
	PublicBaseURL   string
	HTTPClient      *http.Client
	APIURL          string
}

func FromEnvironment(getenv func(string) string) (Config, bool) {
	config := Config{MerchantAccount: getenv("WAYFORPAY_MERCHANT_ACCOUNT"), SecretKey: getenv("WAYFORPAY_SECRET_KEY"), MerchantDomain: getenv("WAYFORPAY_MERCHANT_DOMAIN"), PublicBaseURL: strings.TrimRight(getenv("PUBLIC_BASE_URL"), "/")}
	if config.MerchantAccount == "" || config.SecretKey == "" || config.MerchantDomain == "" || config.PublicBaseURL == "" {
		return Config{}, false
	}
	config.HTTPClient = &http.Client{Timeout: 10 * time.Second}
	config.APIURL = StatusURL
	return config, true
}

type Provider struct {
	config Config
	client *http.Client
}

func New(config Config) *Provider {
	client := config.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &Provider{config: config, client: client}
}

func (provider *Provider) Checkout(_ context.Context, order domain.Order, buyer ports.Buyer, urls ports.CheckoutURLs) (ports.CheckoutForm, error) {
	amount := formatAmount(order.Amount())
	orderDate := strconv.FormatInt(order.CreatedAt().Unix(), 10)
	name := urls.ProductName
	if name == "" {
		name = order.CourseID()
	}
	values := []string{provider.config.MerchantAccount, provider.config.MerchantDomain, order.ID(), orderDate, amount, string(order.Amount().Currency()), name, "1", amount}
	fields := []ports.CheckoutField{
		{Name: "merchantAccount", Value: provider.config.MerchantAccount},
		{Name: "merchantAuthType", Value: "SimpleSignature"},
		{Name: "merchantDomainName", Value: provider.config.MerchantDomain},
		{Name: "merchantSignature", Value: provider.signature(values...)},
		{Name: "orderReference", Value: order.ID()},
		{Name: "orderDate", Value: orderDate},
		{Name: "amount", Value: amount},
		{Name: "currency", Value: string(order.Amount().Currency())},
		{Name: "productName[]", Value: name},
		{Name: "productCount[]", Value: "1"},
		{Name: "productPrice[]", Value: amount},
		{Name: "returnUrl", Value: urls.ReturnURL},
		{Name: "serviceUrl", Value: urls.ServiceURL},
		{Name: "language", Value: "UA"},
	}
	if buyer.Email != "" {
		fields = append(fields, ports.CheckoutField{Name: "clientEmail", Value: buyer.Email})
	}
	return ports.CheckoutForm{Action: PayURL, Fields: fields}, nil
}

func (provider *Provider) OrderStatus(ctx context.Context, orderID string) (ports.PaymentStatus, error) {
	request := statusRequest{TransactionType: "CHECK_STATUS", MerchantAccount: provider.config.MerchantAccount, OrderReference: orderID, APIVersion: 1}
	request.MerchantSignature = provider.signature(provider.config.MerchantAccount, orderID)
	body, err := json.Marshal(request)
	if err != nil {
		return ports.PaymentStatus{}, err
	}
	endpoint := provider.config.APIURL
	if endpoint == "" {
		endpoint = StatusURL
	}
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return ports.PaymentStatus{}, err
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	response, err := provider.client.Do(httpRequest)
	if err != nil {
		return ports.PaymentStatus{}, err
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, maxPayload+1))
	if err != nil || len(responseBody) > maxPayload || response.StatusCode < 200 || response.StatusCode >= 300 {
		return ports.PaymentStatus{}, errors.New("wayforpay status unavailable")
	}
	var payload callbackPayload
	if err := json.Unmarshal(responseBody, &payload); err != nil {
		return ports.PaymentStatus{}, err
	}
	event, err := provider.event(payload, time.Now())
	if err != nil {
		return ports.PaymentStatus{}, err
	}
	return ports.PaymentStatus{Event: event}, nil
}

func (provider *Provider) VerifyCallback(body []byte, receivedAt time.Time) (ports.Callback, error) {
	if len(body) == 0 || len(body) > maxPayload || receivedAt.IsZero() {
		return ports.Callback{}, errors.New("invalid callback")
	}
	var payload callbackPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return ports.Callback{}, errors.New("invalid callback")
	}
	if payload.MerchantAccount != provider.config.MerchantAccount || !provider.validCallbackSignature(payload) {
		return ports.Callback{}, errors.New("invalid callback signature")
	}
	if payload.TransactionStatus != "Approved" && payload.TransactionStatus != "Declined" && payload.TransactionStatus != "Expired" {
		return ports.Callback{OrderReference: payload.OrderReference, ReceivedAt: receivedAt}, nil
	}
	event, err := provider.event(payload, receivedAt)
	if err != nil {
		return ports.Callback{}, err
	}
	return ports.Callback{Event: &event, OrderReference: payload.OrderReference, ReceivedAt: receivedAt}, nil
}

func (provider *Provider) Acknowledge(orderReference string, now time.Time) []byte {
	timestamp := strconv.FormatInt(now.Unix(), 10)
	response := acknowledgement{OrderReference: orderReference, Status: "accept", Time: now.Unix()}
	response.Signature = provider.signature(orderReference, response.Status, timestamp)
	encoded, _ := json.Marshal(response)
	return encoded
}

type callbackPayload struct {
	MerchantAccount   string          `json:"merchantAccount"`
	OrderReference    string          `json:"orderReference"`
	MerchantSignature string          `json:"merchantSignature"`
	Amount            json.RawMessage `json:"amount"`
	Currency          string          `json:"currency"`
	AuthCode          string          `json:"authCode"`
	CardPan           string          `json:"cardPan"`
	TransactionStatus string          `json:"transactionStatus"`
	ReasonCode        json.RawMessage `json:"reasonCode"`
}

type statusRequest struct {
	TransactionType   string `json:"transactionType"`
	MerchantAccount   string `json:"merchantAccount"`
	OrderReference    string `json:"orderReference"`
	MerchantSignature string `json:"merchantSignature"`
	APIVersion        int    `json:"apiVersion"`
}
type acknowledgement struct {
	OrderReference string `json:"orderReference"`
	Status         string `json:"status"`
	Time           int64  `json:"time"`
	Signature      string `json:"signature"`
}

func (provider *Provider) event(payload callbackPayload, receivedAt time.Time) (domain.PaymentEvent, error) {
	amount, err := parseAmount(rawText(payload.Amount), payload.Currency)
	if err != nil || payload.OrderReference == "" || receivedAt.IsZero() {
		return domain.PaymentEvent{}, errors.New("invalid payment status")
	}
	typeValue := domain.PaymentEventType("")
	switch payload.TransactionStatus {
	case "Approved":
		typeValue = domain.PaymentSucceeded
	case "Declined", "Expired":
		typeValue = domain.PaymentFailed
	default:
		return domain.PaymentEvent{}, nil
	}
	return domain.PaymentEvent{Provider: ProviderName, EventID: payload.OrderReference + ":" + payload.TransactionStatus, OrderID: payload.OrderReference, Type: typeValue, Amount: amount, ReceivedAt: receivedAt}, nil
}

func (provider *Provider) validCallbackSignature(payload callbackPayload) bool {
	want := provider.signature(payload.MerchantAccount, payload.OrderReference, rawText(payload.Amount), payload.Currency, payload.AuthCode, payload.CardPan, payload.TransactionStatus, rawText(payload.ReasonCode))
	return constantTimeHexEqual(want, payload.MerchantSignature)
}

func (provider *Provider) signature(values ...string) string {
	mac := hmac.New(md5.New, []byte(provider.config.SecretKey))
	_, _ = mac.Write([]byte(strings.Join(values, ";")))
	return hex.EncodeToString(mac.Sum(nil))
}
func constantTimeHexEqual(left, right string) bool {
	if len(left) != len(right) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(strings.ToLower(left)), []byte(strings.ToLower(right))) == 1
}
func formatAmount(money domain.Money) string {
	return fmt.Sprintf("%d.%02d", money.MinorUnits()/100, money.MinorUnits()%100)
}

func rawText(value json.RawMessage) string {
	text := strings.TrimSpace(string(value))
	if len(text) >= 2 && text[0] == '"' && text[len(text)-1] == '"' {
		var decoded string
		if json.Unmarshal(value, &decoded) == nil {
			return decoded
		}
	}
	return text
}

func parseAmount(value, currency string) (domain.Money, error) {
	parts := strings.Split(value, ".")
	if len(parts) > 2 || parts[0] == "" || strings.HasPrefix(parts[0], "-") || (len(parts) == 2 && len(parts[1]) > 2) {
		return domain.Money{}, domain.ErrMoneyInvalid
	}
	cents := "00"
	if len(parts) == 2 {
		cents = parts[1]
		if len(cents) == 1 {
			cents += "0"
		}
	}
	minor, err := strconv.ParseInt(parts[0]+cents, 10, 64)
	if err != nil {
		return domain.Money{}, domain.ErrMoneyInvalid
	}
	return domain.NewMoney(strconv.FormatInt(minor, 10), currency)
}

var _ ports.PaymentProvider = (*Provider)(nil)
var _ ports.CallbackVerifier = (*Provider)(nil)
