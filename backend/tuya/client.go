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
	accessToken := os.Getenv("TUYA_ACCESS_TOKEN")

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

func (tc *TuyaClient) signService(method, path string, q url.Values, body []byte) (sign, t string) {
	h := sha256.New()
	h.Write(body)
	contentSHA := hex.EncodeToString(h.Sum(nil))

	headersStr := ""

	urlStr := path
	if q != nil {
		qs := q.Encode()
		if qs != "" {
			urlStr += "?" + qs
		}
	}

	stringToSign := strings.Join([]string{
		strings.ToUpper(method),
		contentSHA,
		headersStr,
		urlStr,
	}, "\n")

	t = strconv.FormatInt(time.Now().UTC().UnixMilli(), 10)
	nonce := ""

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

	sign, t := tc.signService(http.MethodGet, path, nil, []byte{})

	req.Header.Set("client_id", tc.ClientID)
	if tc.AccessToken != "" {
		req.Header.Set("access_token", tc.AccessToken)
	}
	req.Header.Set("t", t)
	req.Header.Set("sign", sign)
	req.Header.Set("sign_method", "HMAC-SHA256")

	resp, err := tc.httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	return resp.StatusCode, body, nil
}
