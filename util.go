package emailverifier

import (
	"crypto/md5"
	"encoding/hex"
	"reflect"
	"strings"

	"golang.org/x/net/idna"
)

// parsedDomain parses and returns second level domain
func parsedDomain(domain string) string {
	lowercaseDomain := strings.ToLower(domain)
	parts := strings.Split(lowercaseDomain, ".")
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "." + parts[len(parts)-1]
	}
	return lowercaseDomain
}

// domainToASCII converts any internationalized domain names to ASCII
// reference: https://en.wikipedia.org/wiki/Punycode
func domainToASCII(domain string) string {
	asciiDomain, err := idna.ToASCII(domain)
	if err != nil {
		return domain
	}
	return asciiDomain

}

// callJobFuncWithParams convert jobFunc and prams to a specific function and call it
func callJobFuncWithParams(jobFunc interface{}, params []interface{}) []reflect.Value {
	typ := reflect.TypeOf(jobFunc)
	if typ.Kind() != reflect.Func {
		return nil
	}
	f := reflect.ValueOf(jobFunc)
	if len(params) != f.Type().NumIn() {
		return nil
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	return f.Call(in)
}

// getMD5Hash use md5 to encode string
// #nosec
func getMD5Hash(str string) (error, string) {
	h := md5.New()
	_, err := h.Write([]byte(str))
	if err != nil {
		return err, ""
	}
	return nil, hex.EncodeToString(h.Sum(nil))
}
