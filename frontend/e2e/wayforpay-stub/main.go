package main

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

type order struct {
	Amount   string
	Currency string
}

type stub struct {
	merchant    string
	secret      string
	mu          sync.Mutex
	orders      map[string]order
	payCount    int
	statusCount int
}

type statusRequest struct {
	TransactionType   string `json:"transactionType"`
	MerchantAccount   string `json:"merchantAccount"`
	OrderReference    string `json:"orderReference"`
	MerchantSignature string `json:"merchantSignature"`
	APIVersion        int    `json:"apiVersion"`
}

func main() {
	merchant := os.Getenv("WAYFORPAY_MERCHANT_ACCOUNT")
	secret := os.Getenv("WAYFORPAY_SECRET_KEY")
	if merchant == "" || secret == "" {
		panic("WAYFORPAY_MERCHANT_ACCOUNT and WAYFORPAY_SECRET_KEY are required")
	}
	service := &stub{merchant: merchant, secret: secret, orders: make(map[string]order)}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("GET /debug", service.debug)
	mux.HandleFunc("POST /pay", service.pay)
	mux.HandleFunc("POST /api", service.status)
	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}

func (service *stub) pay(w http.ResponseWriter, request *http.Request) {
	if err := request.ParseForm(); err != nil {
		http.Error(w, "invalid checkout", http.StatusBadRequest)
		return
	}
	orderReference := request.Form.Get("orderReference")
	returnURL := request.Form.Get("returnUrl")
	amount := request.Form.Get("amount")
	currency := request.Form.Get("currency")
	if orderReference == "" || returnURL == "" || amount == "" || currency == "" {
		http.Error(w, "invalid checkout", http.StatusBadRequest)
		return
	}
	service.mu.Lock()
	service.orders[orderReference] = order{Amount: amount, Currency: currency}
	service.payCount++
	service.mu.Unlock()
	http.Redirect(w, request, returnURL, http.StatusSeeOther)
}

func (service *stub) status(w http.ResponseWriter, request *http.Request) {
	var input statusRequest
	body, err := io.ReadAll(io.LimitReader(request.Body, 64*1024+1))
	if err != nil || len(body) > 64*1024 || json.Unmarshal(body, &input) != nil {
		http.Error(w, "invalid status request", http.StatusBadRequest)
		return
	}
	if input.TransactionType != "CHECK_STATUS" || input.MerchantAccount != service.merchant || input.APIVersion != 1 ||
		!hmac.Equal([]byte(input.MerchantSignature), []byte(sign(service.secret, service.merchant+";"+input.OrderReference))) {
		http.Error(w, "invalid status signature", http.StatusUnauthorized)
		return
	}
	service.mu.Lock()
	known, ok := service.orders[input.OrderReference]
	service.statusCount++
	service.mu.Unlock()
	if !ok {
		http.Error(w, "unknown order", http.StatusNotFound)
		return
	}
	payload := map[string]any{
		"merchantAccount":   service.merchant,
		"orderReference":    input.OrderReference,
		"amount":            known.Amount,
		"currency":          known.Currency,
		"authCode":          "e2e",
		"cardPan":           "44****1212",
		"transactionStatus": "Approved",
		"reasonCode":        "1100",
	}
	encoded, _ := json.Marshal(payload)
	var signed map[string]any
	_ = json.Unmarshal(encoded, &signed)
	signed["merchantSignature"] = sign(service.secret, strings.Join([]string{service.merchant, input.OrderReference, known.Amount, known.Currency, "e2e", "44****1212", "Approved", "1100"}, ";"))
	encoded, _ = json.Marshal(signed)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(encoded)
}

func (service *stub) debug(w http.ResponseWriter, _ *http.Request) {
	service.mu.Lock()
	defer service.mu.Unlock()
	_ = json.NewEncoder(w).Encode(map[string]int{"payCount": service.payCount, "statusCount": service.statusCount})
}

func sign(secret, value string) string {
	mac := hmac.New(md5.New, []byte(secret))
	_, _ = mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}
