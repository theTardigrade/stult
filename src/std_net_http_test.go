package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type stdNetHTTPRunMode struct {
	Name string
	Run  func(source string) error
}

func TestStdNetHTTPRequestTextResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != "POST" {
			t.Fatalf("unexpected method: %s", request.Method)
		}
		if got := request.Header.Get("x-request"); got != "example" {
			t.Fatalf("unexpected request header: %q", got)
		}

		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatalf("could not read request body: %v", err)
		}

		response.Header().Add("x-reply", "one")
		response.Header().Add("x-reply", "two")
		response.WriteHeader(http.StatusCreated)
		_, _ = response.Write([]byte("method=" + request.Method + ";body=" + string(body)))
	}))
	defer server.Close()

	source := fmt.Sprintf(`HTTP : STD.NET.HTTP
ASSERT : STD.ERROR.ASSERT

response : HTTP.REQUEST(%q, {
	.METHOD : "POST"
	.HEADERS : {
		{"x-request", "example"}
	}
	.BODY : "payload"
	.TIMEOUT_MILLI : 5000
	.MAX_BYTES : 1000
})

ASSERT.EQUAL(response.STATUS, 201, "status should be returned")
ASSERT.EQUAL(response.BODY, "method=POST;body=payload", "body should be returned as text")
ASSERT.EQUAL(response.URL, %q, "final URL should be returned")

reply_count : 0
((response.HEADERS)) { (header)
	(header[0] = "X-Reply") {
		@reply_count :+ 1
	}
}
ASSERT.EQUAL(reply_count, 2, "duplicate response headers should be preserved")
`, server.URL, server.URL)

	for _, mode := range stdNetHTTPRunModes() {
		t.Run(mode.Name, func(t *testing.T) {
			if err := mode.Run(source); err != nil {
				t.Fatalf("run failed: %v", err)
			}
		})
	}
}

func TestStdNetHTTPRequestByteResponseAndByteRequestBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != "POST" {
			t.Fatalf("unexpected method: %s", request.Method)
		}

		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatalf("could not read request body: %v", err)
		}
		if len(body) != 2 || body[0] != 0 || body[1] != 255 {
			t.Fatalf("unexpected request body bytes: %#v", body)
		}

		_, _ = response.Write([]byte{0, 255, 65})
	}))
	defer server.Close()

	source := fmt.Sprintf(`HTTP : STD.NET.HTTP
ASSERT : STD.ERROR.ASSERT

response : HTTP.REQUEST(%q, {
	.METHOD : "POST"
	.BODY : {0, 255}
	.USE_BYTES : +
	.MAX_BYTES : 10
})

ASSERT.EQUAL(response.STATUS, 200, "status should be returned")
ASSERT.EQUAL(response.BODY[0], 0, "byte zero should be preserved")
ASSERT.EQUAL(response.BODY[1], 255, "byte 255 should be preserved")
ASSERT.EQUAL(response.BODY[2], 65, "following byte should be preserved")
`, server.URL)

	for _, mode := range stdNetHTTPRunModes() {
		t.Run(mode.Name, func(t *testing.T) {
			if err := mode.Run(source); err != nil {
				t.Fatalf("run failed: %v", err)
			}
		})
	}
}

func TestStdNetHTTPRequestContractsCanBeUsedByUserCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("content-type", "text/plain")
		_, _ = response.Write([]byte("ok"))
	}))
	defer server.Close()

	source := fmt.Sprintf(`HTTP : STD.NET.HTTP
ASSERT : STD.ERROR.ASSERT

options<HTTP.REQUEST_OPTIONS_CONTRACT> : {
	.METHOD : "GET"
	.HEADERS : {
		{"accept", "text/plain"}
	}
	.MAX_BYTES : 100
}

response<HTTP.REQUEST_RESPONSE_CONTRACT> : HTTP.REQUEST(%q, options)

MAKE_REQUEST : { (provided<STD.NET.HTTP.REQUEST_OPTIONS_CONTRACT>) : HTTP.REQUEST_RESPONSE_CONTRACT
	(HTTP.REQUEST(%q, provided))
}

wrapped : MAKE_REQUEST(options)

ASSERT.EQUAL(response.STATUS, 200, "response status should satisfy the response contract")
ASSERT.EQUAL(response.BODY, "ok", "response body should satisfy the response contract")
ASSERT.EQUAL(wrapped.STATUS, 200, "dotted contract aliases should work inside functions")
`, server.URL, server.URL)

	for _, mode := range stdNetHTTPRunModes() {
		t.Run(mode.Name, func(t *testing.T) {
			if err := mode.Run(source); err != nil {
				t.Fatalf("run failed: %v", err)
			}
		})
	}
}

func TestStdNetHTTPRequestDoesNotTreatHTTPStatusAsRuntimeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusNotFound)
		_, _ = response.Write([]byte("missing"))
	}))
	defer server.Close()

	source := fmt.Sprintf(`HTTP : STD.NET.HTTP
ASSERT : STD.ERROR.ASSERT

response : HTTP.REQUEST(%q)

ASSERT.EQUAL(response.STATUS, 404, "HTTP status should be data, not a runtime error")
ASSERT.EQUAL(response.BODY, "missing", "error response bodies should still be returned")
`, server.URL)

	for _, mode := range stdNetHTTPRunModes() {
		t.Run(mode.Name, func(t *testing.T) {
			if err := mode.Run(source); err != nil {
				t.Fatalf("run failed: %v", err)
			}
		})
	}
}

func TestStdNetHTTPRequestCanDisableRedirects(t *testing.T) {
	server := httptest.NewServer(http.NewServeMux())
	mux := server.Config.Handler.(*http.ServeMux)
	mux.HandleFunc("/start", func(response http.ResponseWriter, request *http.Request) {
		http.Redirect(response, request, "/final", http.StatusFound)
	})
	mux.HandleFunc("/final", func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write([]byte("final"))
	})
	defer server.Close()

	source := fmt.Sprintf(`HTTP : STD.NET.HTTP
ASSERT : STD.ERROR.ASSERT

followed : HTTP.REQUEST(%q)
not_followed : HTTP.REQUEST(%q, {
	.FOLLOW_REDIRECTS : -
})

ASSERT.EQUAL(followed.STATUS, 200, "redirects should be followed by default")
ASSERT.EQUAL(followed.BODY, "final", "followed response body should be returned")
ASSERT.EQUAL(followed.URL, %q, "final URL should be returned after redirect")
ASSERT.EQUAL(not_followed.STATUS, 302, "redirects can be disabled")
ASSERT.EQUAL(not_followed.URL, %q, "original URL should be returned when redirects are disabled")
`, server.URL+"/start", server.URL+"/start", server.URL+"/final", server.URL+"/start")

	for _, mode := range stdNetHTTPRunModes() {
		t.Run(mode.Name, func(t *testing.T) {
			if err := mode.Run(source); err != nil {
				t.Fatalf("run failed: %v", err)
			}
		})
	}
}

func TestStdNetHTTPRequestRejectsInvalidOptions(t *testing.T) {
	cases := []struct {
		Name     string
		Source   string
		Expected string
	}{
		{
			Name: "unknown option",
			Source: `STD.NET.HTTP.REQUEST("http://127.0.0.1", {
	.LABEL : "bad"
})`,
			Expected: `does not allow map key "LABEL"`,
		},
		{
			Name: "header shape",
			Source: `STD.NET.HTTP.REQUEST("http://127.0.0.1", {
	.HEADERS : {
		{"content-type"}
	}
})`,
			Expected: "expected a two-string array",
		},
		{
			Name: "body type",
			Source: `STD.NET.HTTP.REQUEST("http://127.0.0.1", {
	.BODY : {:}
})`,
			Expected: "expects string, array of number or void value, got map value",
		},
	}

	for _, testCase := range cases {
		for _, mode := range stdNetHTTPRunModes() {
			t.Run(testCase.Name+"/"+mode.Name, func(t *testing.T) {
				err := mode.Run(testCase.Source)
				if err == nil {
					t.Fatal("expected run to fail")
				}
				if !strings.Contains(err.Error(), testCase.Expected) {
					t.Fatalf("expected error to contain %q, got: %v", testCase.Expected, err)
				}
			})
		}
	}
}

func TestStdNetHTTPRequestMaxBytes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write([]byte("too long"))
	}))
	defer server.Close()

	source := fmt.Sprintf(`STD.NET.HTTP.REQUEST(%q, {
	.MAX_BYTES : 3
})`, server.URL)

	for _, mode := range stdNetHTTPRunModes() {
		t.Run(mode.Name, func(t *testing.T) {
			err := mode.Run(source)
			if err == nil {
				t.Fatal("expected run to fail")
			}
			if !strings.Contains(err.Error(), "response body exceeded MAX_BYTES") {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func stdNetHTTPRunModes() []stdNetHTTPRunMode {
	return []stdNetHTTPRunMode{
		{
			Name: "interpreter",
			Run: func(source string) error {
				return runSourceString(NewInterpreter(), source, "<std-net-http-test>")
			},
		},
		{
			Name: "bytecode",
			Run: func(source string) error {
				return runSourceStringWithBytecodeVM(NewBytecodeVM(nil), source, "<std-net-http-test>")
			},
		},
	}
}
