package auth

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	cases := []struct {
		pw string
		ok bool
	}{
		{
			pw: "password",
			ok: true,
		},
		{
			pw: "1234",
			ok: true,
		},
		{
			pw: "passwordasdfasdfasdfasdfasdfasdfasdfasdfadsfasdfasdfasdfasdfasdfasfdpasswordasdfasdfasdfasdfasdfasdfasdfasdfadsfasdfasdfasdfasdfasdfasfdpasswordasdfasdfasdfasdfasdfasdfasdfasdfadsfasdfasdfasdfasdfasdfasfdpasswordasdfasdfasdfasdfasdfasdfasdfasdfadsfasdfasdfasdfasdfasdfasfdpasswordasdfasdfasdfasdfasdfasdfasdfasdfadsfasdfasdfasdfasdfasdfasfdpasswordasdfasdfasdfasdfasdfasdfasdfasdfadsfasdfasdfasdfasdfasdfasfdpasswordasdfasdfasdfasdfasdfasdfasdfasdfadsfasdfasdfasdfasdfasdfasfdpasswordasdfasdfasdfasdfasdfasdfasdfasdfadsfasdfasdfasdfasdfasdfasfd",
			ok: false,
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("Test case %v", i), func(t *testing.T) {
			hpw, err := HashPassword(c.pw)
			if err != nil && c.ok {
				t.Errorf("expected no errors")
			}
			if err = CheckPasswordHash(hpw, c.pw); err != nil && c.ok {
				t.Errorf("expected password '%s' to succeed check", c.pw)
			}
		})
	}
}

func TestMakeAndValidateJWT(t *testing.T) {
	cases := []struct {
		uuid string
		tokenSecret string
	}{
		{
			uuid: "34277abf-5023-11f0-9fc4-00155d7029f7",
			tokenSecret: "hello",
		},
		{
			uuid: "34277eec-5023-11f0-9fc4-00155d7029f7",
			tokenSecret: "hello",
		},
		{
			uuid: "34278025-5023-11f0-9fc4-00155d7029f7",
			tokenSecret: "hello",
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("Test case %v", i), func(t *testing.T) {
			jwt, err := MakeJWT(uuid.MustParse(c.uuid), c.tokenSecret, time.Hour)
			if err != nil {
				t.Errorf("expected no error making jwt: %v", err)
			}
			if len(jwt) != 208 {
				t.Errorf("expected jwt of length 208: got %v", len(jwt))
			}
			
			resUUID, err := ValidateJWT(jwt, c.tokenSecret)
			if err != nil {
				t.Errorf("expected no error validating jwt: %v", err)
			}
			if resUUID.String() != c.uuid {
				t.Errorf("expected returned uuid of %s: got %s", c.uuid, resUUID)
			}
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	cases := []struct{
		header http.Header
		expected string
	}{
		{
			header: http.Header{
				"Authorization": []string{"Bearer 34277abf-5023-11f0-9fc4-00155d7029f7",},
			},
			expected: "34277abf-5023-11f0-9fc4-00155d7029f7",
		},
		{
			header: http.Header{
				"Authorization": []string{"Bearer 34277eec-5023-11f0-9fc4-00155d7029f7",},
			},
			expected: "34277eec-5023-11f0-9fc4-00155d7029f7",
		},
		{
			header: http.Header{
				"Authorization": []string{"34278025-5023-11f0-9fc4-00155d7029f7",},
			},
			expected: "error: missing 'Bearer ' prefix",
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("Test case %v", i), func(t *testing.T) {
			res, err := GetBearerToken(c.header)
			if err != nil {
				if err.Error() != c.expected {
					t.Errorf("expected the error: '%s' or no error", c.expected)
				}
			} else {
				if res != c.expected {
					t.Errorf("expected %s, got %s", c.expected, res)
				}
			}
		})
	}
}