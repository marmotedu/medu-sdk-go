package sdk

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/marmotedu/component-base/pkg/auth"

	"github.com/marmotedu/medu-sdk-go/sdk/log"

	uuid "github.com/satori/go.uuid"
)

const (
	authHeaderPrefix = "MEDU-HMAC-SHA256"
	timeFormat       = "20060102T150405Z"
	shortTimeFormat  = "20060102"

	// emptyStringSHA256 is a SHA256 of an empty string
	emptyStringSHA256 = `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`
)

var ignoredHeaders = []string{"Authorization", "User-Agent", "X-Request-Id"}
var noEscape [256]bool

func init() {
	for i := 0; i < len(noEscape); i++ {
		// expects every character except these to be escaped
		noEscape[i] = (i >= 'A' && i <= 'Z') ||
			(i >= 'a' && i <= 'z') ||
			(i >= '0' && i <= '9') ||
			i == '-' ||
			i == '.' ||
			i == '_' ||
			i == '~'
	}
}

type Signer interface {
	Sign(serviceName string, r *http.Request, body io.ReadSeeker) http.Header
}

type BaseSignature struct {
	Credentials *Credential
	Logger      log.Logger
}

type SignatureV1 struct {
	BaseSignature
}
type SignatureV2 struct {
	BaseSignature
}

type signingCtx struct {
	ServiceName string
	//Region           string
	Request          *http.Request
	Body             io.ReadSeeker
	Query            url.Values
	Time             time.Time
	ExpireTime       time.Duration
	SignedHeaderVals http.Header

	credValues         *Credential
	formattedTime      string
	formattedShortTime string

	bodyDigest       string
	signedHeaders    string
	canonicalHeaders string
	canonicalString  string
	credentialString string
	stringToSign     string
	signature        string
	authorization    string
}

func NewSigner(signMethod string, cred *Credential, logger log.Logger) Signer {
	switch signMethod {
	case "jwt":
		return NewSignatureV1(cred, logger)
	case "hmac":
		return NewSignatureV2(cred, logger)
	default:
		return NewSignatureV1(cred, logger)
	}
}

func NewSignatureV1(cred *Credential, logger log.Logger) Signer {
	return &SignatureV1{
		BaseSignature: BaseSignature{
			Credentials: cred,
			Logger:      logger,
		},
	}
}

func (v1 SignatureV1) Sign(serviceName string, r *http.Request, body io.ReadSeeker) http.Header {
	tokenString := auth.Sign(v1.Credentials.SecretID, v1.Credentials.SecretKey, "medu-sdk-go", serviceName+".marmotedu.com")
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	return r.Header

}

func NewSignatureV2(cred *Credential, logger log.Logger) Signer {
	return &SignatureV2{
		BaseSignature: BaseSignature{
			Credentials: cred,
			Logger:      logger,
		},
	}
}

// Sign signs the request by using AWS 24 signer algorithm, and adds Authorization header
func (v2 SignatureV2) Sign(serviceName string, r *http.Request, body io.ReadSeeker) http.Header {
	return v2.signWithBody(serviceName, r, body, 0)
}

func (v2 SignatureV2) signWithBody(serviceName string, r *http.Request, body io.ReadSeeker, exp time.Duration) http.Header {
	ctx := &signingCtx{
		Request:     r,
		Body:        body,
		Query:       r.URL.Query(),
		Time:        time.Now(),
		ExpireTime:  exp,
		ServiceName: serviceName,
	}

	for key := range ctx.Query {
		sort.Strings(ctx.Query[key])
	}

	if ctx.isRequestSigned() {
		ctx.Time = time.Now()
	}

	ctx.credValues = v2.Credentials
	ctx.build()

	v2.logSigningInfo(ctx)
	return ctx.SignedHeaderVals
}

const logSignInfoMsg = `DEBUG: Request Signature:
---[ CANONICAL STRING  ]-----------------------------
%s
---[ STRING TO SIGN ]--------------------------------
%s%s
-----------------------------------------------------`

func (v2 *SignatureV2) logSigningInfo(ctx *signingCtx) {
	signedURLMsg := ""
	msg := fmt.Sprintf(logSignInfoMsg, ctx.canonicalString, ctx.stringToSign, signedURLMsg)
	v2.Logger.Infof("%s", msg)
}

func (ctx *signingCtx) build() {
	ctx.buildTime()             // no depends
	ctx.buildNonce()            // no depends
	ctx.buildCredentialString() // no depends
	ctx.buildBodyDigest()

	unsignedHeaders := ctx.Request.Header
	ctx.buildCanonicalHeaders(unsignedHeaders)
	ctx.buildCanonicalString() // depends on canon headers / signed headers
	ctx.buildStringToSign()    // depends on canon string
	ctx.buildSignature()       // depends on string to sign

	parts := []string{
		authHeaderPrefix + " Credential=" + ctx.credValues.SecretKey + "/" + ctx.credentialString,
		"SignedHeaders=" + ctx.signedHeaders,
		"Signature=" + ctx.signature,
	}
	ctx.Request.Header.Set("Authorization", strings.Join(parts, ", "))
}

func (ctx *signingCtx) buildTime() {
	ctx.formattedTime = ctx.Time.UTC().Format(timeFormat)
	ctx.formattedShortTime = ctx.Time.UTC().Format(shortTimeFormat)

	ctx.Request.Header.Set("x-medu-date", ctx.formattedTime)
}

func (ctx *signingCtx) buildNonce() {
	ctx.Request.Header.Set("x-medu-nonce", uuid.NewV4().String())
}

func (ctx *signingCtx) buildCredentialString() {
	ctx.credentialString = strings.Join([]string{
		ctx.formattedShortTime,
		ctx.ServiceName,
		"medu_request",
	}, "/")
}

func (ctx *signingCtx) buildCanonicalHeaders(header http.Header) {
	var headers []string
	headers = append(headers, "host")
	for k, v := range header {
		canonicalKey := http.CanonicalHeaderKey(k)
		if shouldIgnore(canonicalKey, ignoredHeaders) {
			continue // ignored header
		}
		if ctx.SignedHeaderVals == nil {
			ctx.SignedHeaderVals = make(http.Header)
		}

		lowerCaseKey := strings.ToLower(k)
		if _, ok := ctx.SignedHeaderVals[lowerCaseKey]; ok {
			// include additional values
			ctx.SignedHeaderVals[lowerCaseKey] = append(ctx.SignedHeaderVals[lowerCaseKey], v...)
			continue
		}

		headers = append(headers, lowerCaseKey)
		ctx.SignedHeaderVals[lowerCaseKey] = v
	}
	sort.Strings(headers)

	ctx.signedHeaders = strings.Join(headers, ";")

	headerValues := make([]string, len(headers))
	for i, k := range headers {
		if k == "host" {
			if ctx.Request.Host != "" {
				headerValues[i] = "host:" + ctx.Request.Host
			} else {
				headerValues[i] = "host:" + ctx.Request.URL.Host
			}
		} else {
			headerValues[i] = k + ":" +
				strings.Join(ctx.SignedHeaderVals[k], ",")
		}
	}
	stripExcessSpaces(headerValues)
	ctx.canonicalHeaders = strings.Join(headerValues, "\n")
}

func (ctx *signingCtx) buildCanonicalString() {
	uri := getURIPath(ctx.Request.URL)

	ctx.canonicalString = strings.Join([]string{
		ctx.Request.Method,
		uri,
		ctx.Request.URL.RawQuery,
		ctx.canonicalHeaders + "\n",
		ctx.signedHeaders,
		ctx.bodyDigest,
	}, "\n")
}

func (ctx *signingCtx) buildStringToSign() {
	ctx.stringToSign = strings.Join([]string{
		authHeaderPrefix,
		ctx.formattedTime,
		ctx.credentialString,
		hex.EncodeToString(makeSha256([]byte(ctx.canonicalString))),
	}, "\n")
}

func (ctx *signingCtx) buildSignature() {
	secret := ctx.credValues.SecretKey
	date := makeHmac([]byte("MEDU"+secret), []byte(ctx.formattedShortTime))
	//region := makeHmac(date, []byte(ctx.Region))
	service := makeHmac(date, []byte(ctx.ServiceName))
	credentials := makeHmac(service, []byte("medu_request"))
	signature := makeHmac(credentials, []byte(ctx.stringToSign))
	ctx.signature = hex.EncodeToString(signature)
}

func (ctx *signingCtx) buildBodyDigest() {
	var hash string
	if ctx.Body == nil {
		hash = emptyStringSHA256
	} else {
		hash = hex.EncodeToString(makeSha256Reader(ctx.Body))
	}

	ctx.bodyDigest = hash
}

// isRequestSigned returns if the request is currently signed or presigned
func (ctx *signingCtx) isRequestSigned() bool {
	if ctx.Request.Header.Get("Authorization") != "" {
		return true
	}

	return false
}

func makeHmac(key []byte, data []byte) []byte {
	hash := hmac.New(sha256.New, key)
	hash.Write(data)
	return hash.Sum(nil)
}

func makeSha256(data []byte) []byte {
	hash := sha256.New()
	hash.Write(data)
	return hash.Sum(nil)
}

func makeSha256Reader(reader io.ReadSeeker) []byte {
	hash := sha256.New()
	start, _ := reader.Seek(0, 1)
	defer reader.Seek(start, 0)

	io.Copy(hash, reader)
	return hash.Sum(nil)
}

const doubleSpace = "  "

// stripExcessSpaces will rewrite the passed in slice's string values to not
// contain multiple side-by-side spaces.
func stripExcessSpaces(vals []string) {
	var j, k, l, m, spaces int
	for i, str := range vals {
		// Trim trailing spaces
		for j = len(str) - 1; j >= 0 && str[j] == ' '; j-- {
		}

		// Trim leading spaces
		for k = 0; k < j && str[k] == ' '; k++ {
		}
		str = str[k : j+1]

		// Strip multiple spaces.
		j = strings.Index(str, doubleSpace)
		if j < 0 {
			vals[i] = str
			continue
		}

		buf := []byte(str)
		for k, m, l = j, j, len(buf); k < l; k++ {
			if buf[k] == ' ' {
				if spaces == 0 {
					// First space.
					buf[m] = buf[k]
					m++
				}
				spaces++
			} else {
				// End of multiple spaces.
				spaces = 0
				buf[m] = buf[k]
				m++
			}
		}

		vals[i] = string(buf[:m])
	}
}

func getURIPath(u *url.URL) string {
	var uri string

	if len(u.Opaque) > 0 {
		uri = "/" + strings.Join(strings.Split(u.Opaque, "/")[3:], "/")
	} else {
		uri = u.EscapedPath()
	}

	if len(uri) == 0 {
		uri = "/"
	}

	return uri
}

func shouldIgnore(header string, ignoreHeaders []string) bool {
	for _, v := range ignoreHeaders {
		if v == header {
			return true
		}
	}
	return false
}

// EscapePath escapes part of a URL path
func EscapePath(path string, encodeSep bool) string {
	var buf bytes.Buffer
	for i := 0; i < len(path); i++ {
		c := path[i]
		if noEscape[c] || (c == '/' && !encodeSep) {
			buf.WriteByte(c)
		} else {
			fmt.Fprintf(&buf, "%%%02X", c)
		}
	}
	return buf.String()
}
