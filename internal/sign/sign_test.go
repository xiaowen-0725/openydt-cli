package sign

import "testing"

// Anchor vectors captured from the platform docs and a real test-env request.
func TestV2Sign(t *testing.T) {
	// Real anchor: MD5("test:20260529220831:123456")
	got := V2Sign("test", "123456", "20260529220831")
	want := "a2710e9f093f22691ff24cf944aadd48"
	if got != want {
		t.Fatalf("V2Sign = %s, want %s", got, want)
	}
}

func TestV3Sign(t *testing.T) {
	// Doc vector: MD5(`abc:20170416142030:{"parkCode":"123123"}:123456`)
	body := CompactBody(`{"parkCode":"123123"}`)
	got := V3Sign("abc", "123456", "20170416142030", body)
	want := "f1e1e7f8bc710dd6633bc0d9a9336207"
	if got != want {
		t.Fatalf("V3Sign = %s, want %s", got, want)
	}
}

func TestAuthorization(t *testing.T) {
	got := Authorization("test", "20260529220831")
	want := "dGVzdDoyMDI2MDUyOTIyMDgzMQ=="
	if got != want {
		t.Fatalf("Authorization = %s, want %s", got, want)
	}
}

func TestCompactBody(t *testing.T) {
	cases := map[string]string{
		"":                                   "{}",
		"{\n  \"parkCode\": \"123123\"\n}":    `{"parkCode":"123123"}`,
		`{"t":"2019-04-16 00:11:25"}`:         `{"t":"2019-04-16 00:11:25"}`, // timestamp space preserved
		"{ \"a\" : 1 , \"b\" : [ 1, 2 ] }":    `{"a":1,"b":[1,2]}`,
	}
	for in, want := range cases {
		if got := CompactBody(in); got != want {
			t.Errorf("CompactBody(%q) = %q, want %q", in, got, want)
		}
	}
}
