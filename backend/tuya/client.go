package tuya

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type TuyaClient struct {
	BaseURL      string
	DeviceID     string
	ClientID     string
	ClientSecret string
	AccessToken  string

	httpClient *http.Client
}

func NewTuyaClientFromEnv() (*TuyaClient, error) {
	baseURL := strings.TrimRight(os.Getenv("TUYA_BASE_URL"), "/")
	deviceID := os.Getenv("TUYA_DEVICE_ID")
	clientID := os.Getenv("TUYA_CLIENT_ID")
	clientSecret := os.Getenv("TUYA_CLIENT_SECRET")
	accessToken := os.Getenv("TUYA_ACCESS_TOKEN") // token de negócio já obtido

	if baseURL == "" || deviceID == "" || clientID == "" || clientSecret == "" {
		return nil, errors.New("TUYA_BASE_URL, TUYA_DEVICE_ID, TUYA_CLIENT_ID e TUYA_CLIENT_SECRET são obrigatórios")
	}

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxConnsPerHost:     32,
		MaxIdleConns:        64,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   12 * time.Second,
	}

	return &TuyaClient{
		BaseURL:      baseURL,
		DeviceID:     deviceID,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AccessToken:  accessToken,
		httpClient:   client,
	}, nil
}

// assina no padrão "service management":
// sign = HMAC-SHA256(client_id + access_token + t + nonce + stringToSign, client_secret)
func (tc *TuyaClient) signService(method, path string, q url.Values, body []byte) (sign, t string) {
	// SHA256 do corpo (GET => corpo vazio)
	h := sha256.New()
	h.Write(body) // nil/[]byte{} => hash do vazio
	contentSHA := hex.EncodeToString(h.Sum(nil)) // para vazio = e3b0c4...b855 (conforme docs)

	// Sem headers adicionais na assinatura (Signature-Headers), então string vazia
	headersStr := ""

	// URL = path + (query ordenada, se existir)
	urlStr := path
	if q != nil {
		qs := q.Encode() // já sai em ordem lexicográfica
		if qs != "" {
			urlStr += "?" + qs
		}
	}

	// stringToSign
	stringToSign := strings.Join([]string{
		strings.ToUpper(method),
		contentSHA,
		headersStr,
		urlStr,
	}, "\n")

	t = strconv.FormatInt(time.Now().UTC().UnixMilli(), 10)
	// nonce opcional (vamos deixar vazio)
	nonce := ""

	// str = client_id + access_token + t + nonce + stringToSign
	base := tc.ClientID + tc.AccessToken + t + nonce + stringToSign

	mac := hmac.New(sha256.New, []byte(tc.ClientSecret))
	mac.Write([]byte(base))
	sign = strings.ToUpper(hex.EncodeToString(mac.Sum(nil)))
	return sign, t
}

func (tc *TuyaClient) GetDevice(ctx context.Context) (int, []byte, error) {
	path := "/v1.0/devices/" + tc.DeviceID
	full := tc.BaseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, full, nil)
	if err != nil {
		return 0, nil, err
	}

	// calcula sign/t p/ esta requisição
	sign, t := tc.signService(http.MethodGet, path, nil, []byte{})

	// headers obrigatórios
	req.Header.Set("client_id", tc.ClientID)
	if tc.AccessToken != "" {
		req.Header.Set("access_token", tc.AccessToken)
	}
	req.Header.Set("t", t)
	req.Header.Set("sign", sign)
	req.Header.Set("sign_method", "HMAC-SHA256")
	// Se algum dia você incluir headers custom na assinatura, também envie:
	// req.Header.Set("Signature-Headers", "header1:header2")

	resp, err := tc.httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	return resp.StatusCode, body, nil
}
