package session

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUserAddr(t *testing.T) {
	tests := []struct {
		name   string
		input  http.Request
		expect string
	}{
		{
			name: "Test0_X-Forwarded-For",
			input: http.Request{
				RemoteAddr: "1.2.3.4:1234",
				Header: http.Header{
					"X-Forwarded-For": []string{"203.0.113.195"},
				},
			},
			expect: "203.0.113.195",
		},
		{
			name: "Test1_X-Forwarded-For",
			input: http.Request{
				RemoteAddr: "1.2.3.4:1234",
				Header: http.Header{
					"X-Forwarded-For": []string{"2001:db8:85a3:8d3:1319:8a2e:370:7348"},
				},
			},
			expect: "2001:db8:85a3:8d3:1319:8a2e:370:7348",
		},
		{
			name: "Test2_X-Forwarded-For",
			input: http.Request{
				RemoteAddr: "1.2.3.4:1234",
				Header: http.Header{
					"X-Forwarded-For": []string{"203.0.113.195, 2001:db8:85a3:8d3:1319:8a2e:370:7348"},
				},
			},
			expect: "203.0.113.195",
		},
		{
			name: "Test3_X-Forwarded-For",
			input: http.Request{
				RemoteAddr: "1.2.3.4:1234",
				Header: http.Header{
					"X-Forwarded-For": []string{"203.0.113.195,2001:db8:85a3:8d3:1319:8a2e:370:7348,198.51.100.178"},
				},
			},
			expect: "203.0.113.195",
		},
		{
			name: "Test0_Forwarded",
			input: http.Request{
				RemoteAddr: "1.2.3.4:1234",
				Header: http.Header{
					"Forwarded": []string{`for="_mdn"`},
				},
			},
			expect: "_mdn",
		},
		{
			name: "Test1_Forwarded",
			input: http.Request{
				RemoteAddr: "1.2.3.4:1234",
				Header: http.Header{
					"Forwarded": []string{`For="[2001:db8:cafe::17]:4711"`},
				},
			},
			expect: "2001:db8:cafe::17",
		},
		{
			name: "Test2_Forwarded",
			input: http.Request{
				RemoteAddr: "1.2.3.4:1234",
				Header: http.Header{
					"Forwarded": []string{`for=192.0.2.60;proto=http;by=203.0.113.43`},
				},
			},
			expect: "192.0.2.60",
		},
		{
			name: "Test3_Forwarded",
			input: http.Request{
				RemoteAddr: "1.2.3.4:1234",
				Header: http.Header{
					"Forwarded": []string{`for=192.0.2.43, for=198.51.100.17`},
				},
			},
			expect: "192.0.2.43",
		},
		{
			name: "Test0_Mixed",
			input: http.Request{
				RemoteAddr: "1.2.3.4:1234",
				Header: http.Header{
					"Forwarded":       []string{`for=192.0.2.172`},
					"X-Forwarded-For": []string{"192.0.2.172"},
				},
			},
			expect: "192.0.2.172",
		},
		{
			name: "Test1_Mixed",
			input: http.Request{
				RemoteAddr: "1.2.3.4:1234",
				Header: http.Header{
					"Forwarded":       []string{`for=192.0.2.43, for="[2001:db8:cafe::17]"`},
					"X-Forwarded-For": []string{"192.0.2.43, 2001:db8:cafe::17"},
				},
			},
			expect: "192.0.2.43",
		},
		{
			name: "Test0_RemoteAddr",
			input: http.Request{
				RemoteAddr: "1.2.3.4:1234",
			},
			expect: "1.2.3.4",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			addr := GetUserAddr(&test.input)
			assert.Exactly(t, test.expect, addr)
		})
	}
}
