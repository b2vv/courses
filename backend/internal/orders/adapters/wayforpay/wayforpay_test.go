package wayforpay

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"courses/internal/orders/domain"
)

func TestPurchaseSignatureMatchesWayForPayWikiVector(t *testing.T) {
	provider := New(Config{SecretKey: "dhkq3vUi94{Z!5frxs(02ML"})
	got := provider.signature("test_merchant", "www.market.ua", "DH783023", "1415379863", "1547.36", "UAH", "Процесор Intel Core i5-4670 3.4GHz", "Пам'ять Kingston DDR3-1600 4096MB PC3-12800", "1", "1", "1000", "547.36")
	if got != "4941ea0c7f5b4833c2bd06b1fefcdc0b" {
		t.Fatalf("signature=%q", got)
	}
}

func TestCallbackVerificationMapsTerminalStatusesAndRejectsTampering(t *testing.T) {
	provider := New(Config{MerchantAccount: "test_merchant", SecretKey: "secret"})
	for _, test := range []struct {
		status string
		want   domain.PaymentEventType
	}{
		{status: "Approved", want: domain.PaymentSucceeded},
		{status: "Declined", want: domain.PaymentFailed},
		{status: "Expired", want: domain.PaymentFailed},
	} {
		t.Run(test.status, func(t *testing.T) {
			body := callbackJSON("test_merchant", "order-1", "1890.00", "UAH", "123", "44****1212", test.status, "1100", "secret")
			callback, err := provider.VerifyCallback(body, time.Unix(100, 0))
			if err != nil || callback.Event == nil || callback.Event.Type != test.want || callback.Event.EventID != "order-1:"+test.status {
				t.Fatalf("callback=%+v error=%v", callback, err)
			}
		})
	}
	tampered := strings.Replace(string(callbackJSON("test_merchant", "order-1", "1890.00", "UAH", "123", "44****1212", "Approved", "1100", "secret")), `"amount":"1890.00"`, `"amount":"1891.00"`, 1)
	if _, err := provider.VerifyCallback([]byte(tampered), time.Unix(100, 0)); err == nil {
		t.Fatal("tampered amount accepted")
	}
	wrongKey := callbackJSON("test_merchant", "order-1", "1890.00", "UAH", "123", "44****1212", "Approved", "1100", "wrong")
	if _, err := provider.VerifyCallback(wrongKey, time.Unix(100, 0)); err == nil {
		t.Fatal("wrong key accepted")
	}
	ignored := callbackJSON("test_merchant", "order-1", "1890.00", "UAH", "", "", "InProcessing", "0", "secret")
	callback, err := provider.VerifyCallback(ignored, time.Unix(100, 0))
	if err != nil || callback.Event != nil || callback.OrderReference != "order-1" {
		t.Fatalf("ignored callback=%+v error=%v", callback, err)
	}
}

func TestAcknowledgeUsesWikiResponseSignature(t *testing.T) {
	provider := New(Config{SecretKey: "secret"})
	got := provider.Acknowledge("order-1", time.Unix(1415379863, 0))
	var response struct {
		OrderReference string `json:"orderReference"`
		Status         string `json:"status"`
		Time           int64  `json:"time"`
		Signature      string `json:"signature"`
	}
	if err := json.Unmarshal(got, &response); err != nil {
		t.Fatal(err)
	}
	if response.Signature != sign("secret", "order-1;accept;1415379863") {
		t.Fatalf("signature=%q", response.Signature)
	}
}

func TestCheckStatusSignsRequestAndVerifiesResponse(t *testing.T) {
	var request statusRequest
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		payload := callbackJSON("test_merchant", "order-1", "1890.00", "UAH", "123", "44****1212", "Approved", "1100", "secret")
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(string(payload))), Header: make(http.Header)}, nil
	})}
	provider := New(Config{MerchantAccount: "test_merchant", SecretKey: "secret", APIURL: "https://api.example.test", HTTPClient: client})
	status, err := provider.OrderStatus(context.Background(), "order-1")
	if err != nil || status.Event.Type != domain.PaymentSucceeded {
		t.Fatalf("status=%+v error=%v", status, err)
	}
	if request.TransactionType != "CHECK_STATUS" || request.MerchantSignature != sign("secret", "test_merchant;order-1") {
		t.Fatalf("request=%+v", request)
	}
}

func callbackJSON(account, order, amount, currency, auth, pan, status, reason, key string) []byte {
	signature := sign(key, strings.Join([]string{account, order, amount, currency, auth, pan, status, reason}, ";"))
	return []byte(`{"merchantAccount":"` + account + `","orderReference":"` + order + `","merchantSignature":"` + signature + `","amount":"` + amount + `","currency":"` + currency + `","authCode":"` + auth + `","cardPan":"` + pan + `","transactionStatus":"` + status + `","reasonCode":"` + reason + `"}`)
}

func sign(key, value string) string {
	mac := hmac.New(md5.New, []byte(key))
	_, _ = mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }
