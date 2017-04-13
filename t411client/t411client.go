package t411client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	ErrNoToken       = errors.New("no token")
	ErrURLParsing    = errors.New("url parsing error")
	ErrWrongPassword = &errAPI{
		Code: 107,
		Text: "Wrong password",
	}
	ErrAccountDisabled = &errAPI{
		Code: 103,
		Text: "Account is disabled",
	}
	ErrUserNotFound = &errAPI{
		Code: 101,
		Text: "User not found",
	}
	ErrTokenExpired = &errAPI{
		Code: 201,
		Text: "Token has expired. Please login",
	}
	tokenAttempt = "token retrieved, try again"
)

// the base url can change every time the t411 api moves from provider
const (
	t411BaseURL = "https://api.t411.li"
	authAPI     = "/auth"
	// UserAgent is the user agent header used in http requests.
	// You can override it if wanted when using t411client package.
	UserAgent    = "TBotAgent"
	defaultDelay = 24 * 7 * 12 // 9 weeks
)

type errAPI struct {
	Code int    `json:"code"`
	Text string `json:"error"`
}

func (e *errAPI) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Text)
}

// Credentials is the couple username password required by t411 API for authentification
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// token is a struct return by the T411 API when requesting a token
type token struct {
	UID   string `json:"uid"`
	Token string `json:"token"`
}

// T411 represents the web client to the t411 API
type T411 struct {
	baseURL      string
	token        *token
	credentials  Credentials
	httpClient   *http.Client
	maxDelay     float64
	keepRatio    bool
	onlyVerified bool
}

// GetToken returns the token retrieved from authentication, if any.
func (t *T411) GetToken() (string, error) {
	if t.token != nil {
		return t.token.Token, nil
	}
	return "", ErrNoToken
}

// SetMaxDelay sets the maximum delay allowed to have between
// the release date of a show episode and the added date of a torrent
// in the t411 tracker.
func (t *T411) SetMaxDelay(maxDelay float64) {
	t.maxDelay = maxDelay
}

// GetMaxDelay returns the maximum delay allowed to have between
// the release date of a show episode and the added date of a torrent
// in the t411 tracker.
func (t *T411) GetMaxDelay() float64 {
	return t.maxDelay
}

// KeepRatio disables any download that could put the ratio below 1.
// By default, keepRatio is set to true.
func (t *T411) KeepRatio(keepRatio bool) {
	t.keepRatio = keepRatio
}

// OnlyVerified enables download only of verified torrents.
// By default, onlyVerified is set to false.
func (t *T411) OnlyVerified(onlyVerified bool) {
	t.onlyVerified = onlyVerified
}

func newEmptyClient(baseURL, username, password string) *T411 {
	if len(baseURL) == 0 {
		baseURL = t411BaseURL
	}
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	return &T411{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout:   time.Second * 10,
			Transport: netTransport,
		},
		credentials: Credentials{
			Username: username,
			Password: password,
		},
		token:        &token{},
		maxDelay:     defaultDelay,
		keepRatio:    true,
		onlyVerified: false,
	}
}

// NewT411Client creates a T411 web client.
// Note: 'baseURL' is set to the default one if left empty.
// This parameter will be useful when the baseURL of t411 API becomes unavailable.
func NewT411Client(baseURL, username, password string) (*T411, error) {
	client := newEmptyClient(baseURL, username, password)
	err := client.retrieveToken()
	if err != nil {
		return nil, err
	}
	return client, nil
}

// NewT411ClientWithToken creates a T411 web client the same way NewT411Client does
// but with a token parameter from a previous session. If the token is in invalid format it returns an error.
func NewT411ClientWithToken(baseURL, username, password, previousToken string) (*T411, error) {
	if len(previousToken) == 0 {
		return NewT411Client(baseURL, username, password)
	}
	splitted := strings.Split(previousToken, ":")
	if len(splitted) != 3 {
		return nil, fmt.Errorf("%s", "invalid token format, must be of the form 12345:123:abcdefghijklmnopqr")
	}
	client := newEmptyClient(baseURL, username, password)
	client.token.Token = previousToken
	client.token.UID = splitted[0]
	_, err := client.UsersProfile(client.token.UID)
	if err != nil {
		if strings.Contains(err.Error(), tokenAttempt) {
			return client, nil
		}
		return nil, fmt.Errorf("%s", err.Error())
	}
	return client, nil
}

func (t *T411) doRequest(req *http.Request) (*http.Response, error) {
	if len(t.token.Token) != 0 {
		req.Header.Set("Authorization", t.token.Token)
	} else {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", UserAgent)
	return t.httpClient.Do(req)
}

func (t *T411) do(method string, u *url.URL, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	resp, err := t.doRequest(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func fixTorrentList(str, expression, replacement string) (string, bool) {
	re := regexp.MustCompile(expression)
	split := re.Split(str, -1)
	if len(split) >= 2 {
		str = strings.Join(split, replacement)
		return str, true
	}
	return str, false
}

// fixJSONResponse fixes the json response for /torrents/search/ api endpoint for the fields
// offset, limit, total and owner.
func fixJSONResponse(bytes []byte) []byte {
	str := string(bytes)
	str = strings.Replace(str, fmt.Sprintf("%q:0", "offset"), fmt.Sprintf("%q:%q", "offset", "0"), 1)
	str = strings.Replace(str, fmt.Sprintf("%q:10", "limit"), fmt.Sprintf("%q:%q", "limit", "10"), 1)
	str = strings.Replace(str, fmt.Sprintf("%q:0", "total"), fmt.Sprintf("%q:%q", "total", "0"), 1)
	str = strings.Replace(str, fmt.Sprintf("%q:0", "owner"), fmt.Sprintf("%q:%q", "owner", "0"), 1)
	// removes unwanted integers in the list of torrents that appears sometimes
	// and replace them with empty Torrent data.
	str, _ = fixTorrentList(str, ":\\[[0-9]+,", ":[{},")
	str, _ = fixTorrentList(str, ",[0-9]+\\]}", ",{}]}")
	str, _ = fixTorrentList(str, ":\\[[0-9]+\\]}", ":[{}]}")
	for {
		str, ok := fixTorrentList(str, ",[0-9]+,", ",{},")
		if !ok {
			return []byte(str)
		}
	}
}

func decodeErr(resp *http.Response) ([]byte, error) {
	// resp.ContentLength not set properly from server side ?
	buf := bytes.NewBuffer(make([]byte, 0 /*, resp.ContentLength */))
	_, err := buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}
	bytes := buf.Bytes()

	// authentication and other requests can fails with StatusCode = 200...
	// so do not check for the status code value, instead try to
	// unmarshal into an errAPI struct and check if there was an error.
	errorAPI := &errAPI{}
	err = json.Unmarshal(bytes, errorAPI)
	if err != nil {
		return bytes, nil
	}
	if len(errorAPI.Text) != 0 {
		return nil, errorAPI
	}
	return bytes, nil
}

func (t *T411) decode(data interface{}, resp *http.Response, usedAPI, query string) error {
	bytes, err := decodeErr(resp)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(fixJSONResponse(bytes), data); err != nil {
		log.Printf("Error decoding using '%s' API for '%s' query :%v", usedAPI, query, err)
		return err
	}
	return nil
}

// retrieveToken does an authentification request on T411 API
// and retrieve the token needed for further requests.
// Note:the Time-To-Live of the token is 90 days.
func (t *T411) retrieveToken() error {
	usedAPI := authAPI
	u, err := url.Parse(t.baseURL + usedAPI)
	if err != nil {
		return err
	}

	form := url.Values{}
	form.Set("username", t.credentials.Username)
	form.Set("password", t.credentials.Password)
	resp, err := t.do("POST", u, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = t.decode(t.token, resp, usedAPI, "")
	if err != nil {
		return err
	}
	return nil
}
