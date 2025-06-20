package auth

import (
	"fmt"
	"testing"
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