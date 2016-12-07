package t411client

import . "gopkg.in/check.v1"

func (s *MySuite) TestUsersProfile(c *C) {
	t411, username, _ := createT411Client(c)
	user, err := t411.UsersProfile(t411.token.UID)
	c.Assert(err, IsNil)
	c.Assert(user.Username, Equals, username)

	ratio, err := user.GetRatio()
	c.Assert(err, IsNil)
	c.Assert(ratio > 0, Equals, true)
}
