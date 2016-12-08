package t411client

import (
	"strings"

	"strconv"

	"os"

	. "gopkg.in/check.v1"
)

func (s *MySuite) TestMakeURL(c *C) {
	usedAPI, u, err := makeURL("breaking bad", 1, 1, "")
	c.Assert(err, IsNil)
	c.Assert(usedAPI, Equals, "/torrents/search/")
	expected := t411BaseURL + "/torrents/search/breaking%20bad?term%5B45%5D%5B%5D=968&term%5B46%5D%5B%5D=937"
	c.Assert(u.String(), Equals, expected)

	usedAPI, u, err = makeURL("vikings", 1, 1, "")
	c.Assert(err, IsNil)
	c.Assert(usedAPI, Equals, "/torrents/search/")
	expected = t411BaseURL + "/torrents/search/vikings?term%5B45%5D%5B%5D=968&term%5B46%5D%5B%5D=937"
	c.Assert(u.String(), Equals, expected)

	usedAPI, u, err = makeURL("vikings", 2, 3, "")
	c.Assert(err, IsNil)
	c.Assert(usedAPI, Equals, "/torrents/search/")
	expected = t411BaseURL + "/torrents/search/vikings?term%5B45%5D%5B%5D=969&term%5B46%5D%5B%5D=939"
	c.Assert(u.String(), Equals, expected)

	usedAPI, u, err = makeURL("vikings", 2, 3, "english")
	c.Assert(err, IsNil)
	c.Assert(usedAPI, Equals, "/torrents/search/")
	expected = t411BaseURL + "/torrents/search/vikings?term%5B45%5D%5B%5D=969&term%5B46%5D%5B%5D=939&term%5B51%5D%5B%5D=1209"
	c.Assert(u.String(), Equals, expected)
}

func checkTorrents(c *C, torrents *Torrents, query, total string, offset, limit int) {
	c.Assert(torrents.Total, Equals, total)
	c.Assert(torrents.Offset, Equals, offset)
	c.Assert(torrents.Limit, Equals, limit)
	c.Assert(torrents.Torrents, Not(HasLen), 0)
	c.Assert(len(torrents.Torrents) <= 10, Equals, true)
	for _, v := range torrents.Torrents {
		c.Assert(strings.Contains(strings.ToLower(v.String()), query), Equals, true)
	}
}

func (s *MySuite) TestSearchTorrentsByTerms(c *C) {
	t411, _, _ := createT411Client(c)
	torrents, err := t411.SearchTorrentsByTerms("vikings", 1, 1, "")
	c.Assert(err, IsNil)
	checkTorrents(c, torrents, "viking", "12", 0, 10)

	torrents, err = t411.SearchTorrentsByTerms("vikings", 1, 1, "french")
	c.Assert(err, IsNil)
	checkTorrents(c, torrents, "viking", "4", 0, 10)
}

func (s *MySuite) TestSearchTorrentsSortingBySeeders(c *C) {
	t411, _, _ := createT411Client(c)
	torrents, err := t411.SearchTorrentsByTerms("vikings", 1, 1, "")
	c.Assert(err, IsNil)
	torrentsList := torrents.Torrents
	c.Assert(torrentsList, Not(HasLen), 0)
	t411.SortBySeeders(torrentsList)
	c.Assert(torrentsList, Not(HasLen), 0)
	current, err := strconv.Atoi(torrentsList[0].Seeders)
	c.Assert(err, IsNil)
	for _, v := range torrentsList {
		c.Assert(strings.Contains(strings.ToLower(v.String()), "viking"), Equals, true)
		temp, err := strconv.Atoi(v.Seeders)
		c.Assert(err, IsNil)
		c.Assert(current <= temp, Equals, true)
		current = temp
	}
}

func (s *MySuite) TestDownloadTorrentByTerms(c *C) {
	t411, _, _ := createT411Client(c)
	path, err := t411.DownloadTorrentByTerms("vikings", 1, 1, "")
	c.Assert(err, IsNil)
	defer func() {
		c.Assert(os.Remove(path), IsNil)
	}()
	c.Assert(path, Not(HasLen), 0)
	c.Assert(strings.Contains(path, "vikingsS01E01"), Equals, true)
}
