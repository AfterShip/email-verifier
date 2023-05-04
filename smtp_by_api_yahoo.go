package emailverifier

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
)

const (
	SIGNUP_PAGE = "https://login.yahoo.com/account/create?specId=yidReg&lang=en-US&src=&done=https%3A%2F%2Fwww.yahoo.com&display=login"
	SIGNUP_API  = "https://login.yahoo.com/account/module/create?validateField=yid"
	// USER_AGENT Fake one to use in API requests
	USER_AGENT = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.71 Safari/537.36"
)

func newYahooAPIVerifier(client *http.Client) smtpAPIVerifier {
	if client == nil {
		client = http.DefaultClient
	}
	return yahoo{
		client: client,
	}
}

type yahoo struct {
	client *http.Client
}

func (y yahoo) isSupported(host string) bool {
	// FIXME Is this `contains` too lenient?
	return strings.Contains(host, "yahoo")
}

func (y yahoo) check(domain, username string) (*SMTP, error) {
	signUpPageResp, err := y.toSignUpPage()
	if err != nil {
		return nil, err
	}
	cookies := signUpPageResp.Cookies()
	if len(cookies) == 0 {
		return nil, errors.New("yahoo check by api, no cookies")
	}
	defer signUpPageResp.Body.Close()
	acrumb := getAcrumb(cookies)
	if acrumb == "" {
		return nil, errors.New("yahoo check by api, no acrumb")
	}
	resp, err := y.sendValidateRequest(username, acrumb, cookies)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	usernameExists, err := checkUsernameExists(resp)
	if err != nil {
		return nil, err
	}
	return &SMTP{
		HostExists:  true,
		Deliverable: usernameExists,
	}, nil
}

func checkUsernameExists(resp *http.Response) (usernameExists bool, err error) {
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	type errItem struct {
		Name  string `json:"name"`
		Error string `json:"error"`
	}
	type errResp struct {
		Errors []errItem `json:"errors"`
	}
	var res errResp
	if err = json.Unmarshal(respBytes, &res); err != nil {
		return false, err
	}
	for _, item := range res.Errors {
		if item.Name == "yid" && item.Error == "IDENTIFIER_EXISTS" {
			return true, nil
		}
	}
	return false, nil
}

func (y yahoo) sendValidateRequest(username string, acrumb string, cookies []*http.Cookie) (*http.Response, error) {
	data, err := json.Marshal(struct {
		Acrumb string `json:"acrumb"`
		SpecId string `json:"specId"`
		Yid    string `json:"yid"`
	}{
		Acrumb: acrumb,
		SpecId: "yidReg",
		Yid:    username,
	})
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest(http.MethodPost, SIGNUP_API, bytes.NewReader(data))
	for _, c := range cookies {
		request.AddCookie(c)
	}
	request.Header.Add("X-Requested-With", "XMLHttpRequest")
	request.Header.Add("Content-Type", "application/json; charset=UTF-8")
	return y.client.Do(request)
}

func (y yahoo) toSignUpPage() (*http.Response, error) {
	request, err := http.NewRequest(http.MethodGet, SIGNUP_PAGE, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("User-Agent", USER_AGENT)
	return y.client.Do(request)
}

func getAcrumb(cookies []*http.Cookie) string {
	for _, c := range cookies {
		re := regexp.MustCompile(`s=(?P<acrumb>[^;^&]*)`)
		match := re.FindStringSubmatch(c.Value)
		if len(match) > 1 {
			return match[1]
		}
	}
	return ""
}
