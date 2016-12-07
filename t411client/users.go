package t411client

import (
	"net/url"
	"strconv"
)

// User returns a user data
type User struct {
	Username   string `json:"username"`
	Gender     string `json:"gender"`
	Age        string `json:"age"`
	Avatar     string `json:"avatar"`
	Downloaded string `json:"downloaded"`
	Uploaded   string `json:"uploaded"`
}

// GetRatio returns the uploaded/downloaded ratio of the user
func (u *User) GetRatio() (float64, error) {
	downloaded, err := strconv.ParseFloat(u.Downloaded, 64)
	if err != nil {
		return -1, err
	}
	uploaded, err := strconv.ParseFloat(u.Uploaded, 64)
	if err != nil {
		return -1, err
	}
	return uploaded / downloaded, nil
}

// UsersProfile gets the user infos of the user with id 'uid'
func (t *T411) UsersProfile(uid string) (*User, error) {
	usedAPI := "/users/profile"
	u, err := url.Parse(t.baseURL + usedAPI + "/" + url.QueryEscape(uid))
	if err != nil {
		return nil, errURLParsing
	}

	resp, err := t.doGet(u)
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
