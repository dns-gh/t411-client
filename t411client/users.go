package t411client

import (
	"fmt"
	"math"
	"net/url"
	"strconv"
)

// User returns a user data.
type User struct {
	Username   string `json:"username"`
	Gender     string `json:"gender"`
	Age        string `json:"age"`
	Avatar     string `json:"avatar"`
	Downloaded string `json:"downloaded"`
	Uploaded   string `json:"uploaded"`
}

// UsersProfile gets the user infos of the user with id 'uid'.
func (t *T411) UsersProfile(uid string) (*User, error) {
	usedAPI := "/users/profile"
	u, err := url.Parse(fmt.Sprintf("%s%s/%s", t.baseURL, usedAPI, url.QueryEscape(uid)))
	if err != nil {
		return nil, errURLParsing
	}

	resp, err := t.do("GET", u, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	user := &User{}
	err = t.decode(user, resp, usedAPI, u.RawQuery)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetOwnProfile gets the profile of the authenticated user.
func (t *T411) GetOwnProfile() (*User, error) {
	return t.UsersProfile(t.token.UID)
}

// GetRatio returns the uploaded/(downloaded+incoming) ratio of the user.
func (u *User) GetRatio(incoming float64) (float64, error) {
	downloaded, err := strconv.ParseFloat(u.Downloaded, 64)
	if err != nil {
		return -1, err
	}
	if downloaded == 0 {
		return math.MaxFloat64, nil
	}
	uploaded, err := strconv.ParseFloat(u.Uploaded, 64)
	if err != nil {
		return -1, err
	}
	return uploaded / (downloaded + incoming), nil
}

// GetOwnRatio returns the uploaded/(downloaded+incoming) ratio
// of the authenticated user.
func (t *T411) GetOwnRatio(incoming float64) (float64, error) {
	user, err := t.GetOwnProfile()
	if err != nil {
		return 0, err
	}
	return user.GetRatio(incoming)
}
