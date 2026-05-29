// Package sign implements the openydt open-platform request signing scheme.
//
// Verified against the platform's reference web client
// (open-api-front/src/views/DebuggingTools.vue) and live test-environment calls:
//
//	ts            = local time, layout yyyyMMddHHmmss, valid for 10 minutes
//	v2 sign       = lower(md5( key + ":" + ts + ":" + secret ))            // body NOT included
//	v3 sign       = lower(md5( key + ":" + ts + ":" + body + ":" + secret )) // compact body included
//	Authorization = base64( key + ":" + ts )                               // same for v2 and v3
//
// The signed body and the body sent over the wire MUST be byte-identical, so
// callers compact once (CompactBody) and reuse the result for both.
package sign

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// Version selects the signing algorithm.
type Version string

const (
	V2 Version = "v2"
	V3 Version = "v3"
)

// TimeLayout is the platform timestamp format (Go reference time).
const TimeLayout = "20060102150405"

// Now returns the current local timestamp in the platform format.
func Now() string { return time.Now().Format(TimeLayout) }

// CompactBody canonicalizes a JSON body the way the platform expects for v3
// signing: insignificant whitespace removed, key order and string-internal
// spaces (e.g. inside "2019-04-16 00:11:25") preserved. Proper JSON compaction
// keeps spaces that live inside string values, so timestamps survive naturally.
//
// If body is empty it becomes "{}". If body is not valid JSON it is returned
// unchanged (callers validate JSON earlier and surface a friendly error).
func CompactBody(body string) string {
	if len(body) == 0 {
		return "{}"
	}
	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(body)); err != nil {
		return body
	}
	return buf.String()
}

func md5hex(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

// V2Sign computes the v2 signature (body not included).
func V2Sign(key, secret, ts string) string {
	return md5hex(key + ":" + ts + ":" + secret)
}

// V3Sign computes the v3 signature over the already-compacted body.
func V3Sign(key, secret, ts, compactBody string) string {
	return md5hex(key + ":" + ts + ":" + compactBody + ":" + secret)
}

// Authorization builds the Authorization header value.
func Authorization(key, ts string) string {
	return base64.StdEncoding.EncodeToString([]byte(key + ":" + ts))
}

// Compute returns the signature for the given version. compactBody must already
// be the canonical body that will be sent on the wire (use CompactBody).
func Compute(v Version, key, secret, ts, compactBody string) (string, error) {
	switch v {
	case V2, "":
		return V2Sign(key, secret, ts), nil
	case V3:
		return V3Sign(key, secret, ts, compactBody), nil
	default:
		return "", fmt.Errorf("unknown sign version %q (want v2 or v3)", v)
	}
}
