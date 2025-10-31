package tuya

import(
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
	"errors"
	"context"
)

type TuyaClient struct {
	BaseURL      string
	DeviceID     string
	ClientID     string
	AccessToken  string
	Sign         string
	TimestampT   string
	SignMethod   string
	Nonce        string
	StringToSign string

	httpClient *http.Client
}

func NewTuyaClientFromEnv() (*TuyaClient, error) {
	baseURL := strings.TrimRight(os.Getenv("TUYA_BASE_URL"), "/")
	deviceID := os.Getenv("TUYA_DEVICE_ID")
	clientID := os.Getenv("TUYA_CLIENT_ID")
	accessToken := os.Getenv("TUYA_ACCESS_TOKEN")
	sign := os.Getenv("TUYA_SIGN")
	t := os.Getenv("TUYA_TIMESTAMP_T")
	signMethod := os.Getenv("TUYA_SIGN_METHOD")
	nonce := os.Getenv("TUYA_NONCE")
	stringToSign := os.Getenv("TUYA_STRING_TO_SIGN")

	if baseURL == "" || deviceID == "" || clientID == "" {
		return nil, errors.New("TUYA_BASE_URL, TUYA_DEVICE_ID e TUYA_CLIENT_ID são obrigatórios")
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
		AccessToken:  accessToken,
		Sign:         sign,
		TimestampT:   t,
		SignMethod:   signMethod,
		Nonce:        nonce,
		StringToSign: stringToSign,
		httpClient:   client,
	}, nil
}

func (tc *TuyaClient) GetDevice(ctx context.Context) (int, []byte, error) {
	url := tc.BaseURL + "/v1.0/devices/" + tc.DeviceID

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, nil, err
	}

	req.Header.Set("client_id", tc.ClientID)

	if tc.AccessToken != "" {
		req.Header.Set("access_token", tc.AccessToken)
	}
	if tc.Sign != "" {
		req.Header.Set("sign", tc.Sign)
	}
	if tc.TimestampT != "" {
		req.Header.Set("t", tc.TimestampT)
	}
	if tc.SignMethod != "" {
		req.Header.Set("sign_method", tc.SignMethod)
	}
	if tc.Nonce != "" {
		req.Header.Set("nonce", tc.Nonce)
	}

	// if tc.StringToSign != "" {
	// 	req.Header.Set("stringToSign", tc.StringToSign)
	// }

	resp, err := tc.httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20)) // 2MB limit
	return resp.StatusCode, body, nil
}
