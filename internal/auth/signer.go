package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// Signer signs outgoing HTTP requests using HMAC-SHA256.
type Signer struct {
	secret []byte
}

// NewSigner creates a Signer with the given shared secret.
func NewSigner(secret string) *Signer {
	return &Signer{secret: []byte(secret)}
}

// Sign adds HMAC-SHA256 signature headers to the request.
func (s *Signer) Sign(req *http.Request, body []byte) error {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	payload := fmt.Sprintf("%s\n%s\n%s\n%s",
		req.Method, req.URL.RequestURI(), ts, string(body))

	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(payload)) // hash.Hash.Write never errors
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	req.Header.Set("X-Token-Timestamp", ts)
	req.Header.Set("X-Token-Signature", sig)
	return nil
}

// VerifyWebhook verifies an inbound webhook HMAC-SHA256 signature.
func (s *Signer) VerifyWebhook(payload []byte, signature string) bool {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write(payload) // hash.Hash.Write never errors
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}
