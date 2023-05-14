package emailverifier

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

const (
	SIGNUP_PAGE = "https://login.yahoo.com/account/create?specId=yidregsimplified&lang=en-US&src=&done=https%3A%2F%2Fwww.yahoo.com&display=login"
	SIGNUP_API  = "https://login.yahoo.com/account/module/create?validateField=userId"
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
	sessionIndex, err := getSessionIndex(signUpPageResp)
	if err != nil {
		return nil, err
	}
	if acrumb == "" {
		return nil, errors.New("yahoo check by api, no acrumb")
	}
	validateResp, err := y.sendValidateRequest(domain, username, acrumb, sessionIndex, cookies)
	if err != nil {
		return nil, err
	}
	defer validateResp.Body.Close()
	usernameExists, err := checkUsernameExists(validateResp)
	if err != nil {
		return nil, err
	}
	return &SMTP{
		HostExists:  true,
		Deliverable: usernameExists,
	}, nil
}

func getSessionIndex(resp *http.Response) (string, error) {
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(`value="([^"]+)" name="sessionIndex"`)

	// 在响应体中查找匹配项
	match := re.FindSubmatch(respBytes)
	if len(match) > 1 {
		return string(match[1]), nil
	}
	return "", errors.New("yahoo check by api, no sessionIndex")
}

func checkUsernameExists(resp *http.Response) (usernameExists bool, err error) {
	respBytes, err := ioutil.ReadAll(resp.Body)
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
		if item.Name == "userId" && item.Error == "IDENTIFIER_EXISTS" {
			return true, nil
		}
	}
	return false, nil
}

func (y yahoo) sendValidateRequest(domain, username, acrumb, sessionIndex string, cookies []*http.Cookie) (*http.Response, error) {
	data, err := json.Marshal(struct {
		Acrumb       string `json:"acrumb"`
		SpecId       string `json:"specId"`
		Yid          string `json:"userId"`
		SessionIndex string `json:"sessionIndex"`
		YidDomain    string `json:"yidDomain"`
	}{
		Acrumb:       acrumb,
		SpecId:       "yidregsimplified",
		Yid:          username,
		SessionIndex: sessionIndex,
		YidDomain:    domain,
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
