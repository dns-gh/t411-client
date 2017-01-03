package t411client

import (
	. "gopkg.in/check.v1"
)

func checkUser(c *C, t411 *T411, username string) {
	user, err := t411.UsersProfile(t411.token.UID)
	c.Assert(err, IsNil)
	c.Assert(user.Username, Equals, username)
}

func (s *MySuite) TestUsersProfile(c *C) {
	t411, username, _ := createT411Client(c)
	checkUser(c, t411, username)

	user, err := t411.GetOwnProfile()
	c.Assert(err, IsNil)
	c.Assert(user.Username, Equals, username)

	ratio, err := user.GetRatio(0)
	c.Assert(err, IsNil)
	c.Assert(ratio > 0, Equals, true)

	ratio, err = t411.GetOwnRatio(0)
	c.Assert(err, IsNil)
	c.Assert(ratio > 0, Equals, true)
}
