package t411client

import (
	"os"
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func createT411Client(c *C) (*T411, string, string) {
	username, password := os.Getenv("T411_USERNAME"), os.Getenv("T411_PASSWORD")
	c.Assert(username, Not(HasLen), 0)
	c.Assert(password, Not(HasLen), 0)

	t411, err := NewT411Client("", username, password)
	c.Assert(err, IsNil)
	return t411, username, password
}

// export T411_USERNAME=YOUR_USERNAME && export T411_PASSWORD=YOUR_PASSWORD && go test ...t411client -gocheck.vv -test.v -gocheck.f TestNAME
func (s *MySuite) TestNewT411(c *C) {
	t411, username, password := createT411Client(c)
	t411Token, err := t411.GetToken()
	c.Assert(err, IsNil)
	c.Assert(t411Token, Not(HasLen), 0)
	expected := &T411{
		baseURL:    t411BaseURL,
		httpClient: t411.httpClient,
		output:     t411.output,
		credentials: Credentials{
			Username: username,
			Password: password,
		},
		token:    t411.token,
		maxDelay: 336,
	}
	c.Assert(t411, DeepEquals, expected)

	t411, err = NewT411Client("", username, "test")
	c.Assert(err, DeepEquals, errWrongPassword)
	c.Assert(t411, IsNil)

	t411, err = NewT411Client("", "test", "test")
	c.Assert(err, DeepEquals, errAccountDisabled)
	c.Assert(t411, IsNil)

	t411, err = NewT411Client("", "test_not_found", "test")
	c.Assert(err, DeepEquals, errUserNotFound)
	c.Assert(t411, IsNil)

	t411, err = NewT411Client("https://api.t411.test", username, "test")
	c.Assert(err, NotNil)
}
