package t411client

import (
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestReqURL(c *C) {
	req := searchReq{
		Title:   "vikings",
		Season:  1,
		Episode: 1,
	}

	c.Assert(req.Title, Equals, "vikings")
	c.Assert(req.Season, Equals, 1)
	c.Assert(req.Episode, Equals, 1)

	expected := t411BaseURL + "/torrents/search/vikings?term%5B45%5D%5B%5D=968&term%5B46%5D%5B%5D=937"
	c.Assert(req.URL(), Equals, expected)

	req = searchReq{
		Title:   "vikings",
		Season:  3,
		Episode: 5,
	}
	expected = t411BaseURL + "/torrents/search/vikings?term%5B45%5D%5B%5D=970&term%5B46%5D%5B%5D=941"
	c.Assert(req.URL(), Equals, expected)
}
