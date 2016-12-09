package t411client

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
)

var (
	errEOF = errors.New("no more torrents to find")
)

// Torrent represents a torrent as return by the t411 API
type Torrent struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Category       string `json:"category"`
	Rewritename    string `json:"rewritename"`
	Seeders        string `json:"seeders"`
	Leechers       string `json:"leechers"`
	Comments       string `json:"comments"`
	IsVerified     string `json:"isVerified"`
	Added          string `json:"added"`
	Size           string `json:"size"`
	TimesCompleted string `json:"times_completed"`
	Owner          string `json:"owner"`
	Categoryname   string `json:"categoryname"`
	Categoryimage  string `json:"categoryimage"`
	Username       string `json:"username"`
	Privacy        string `json:"privacy"`
}

// Torrents represents the torrents data.
type Torrents struct {
	Query    string    `json:"query"`
	Total    string    `json:"total"`
	Offset   string    `json:"offset"`
	Limit    string    `json:"limit"`
	Torrents []Torrent `json:"torrents"`
}

func (t *Torrent) String() string {
	return fmt.Sprintf("%s - %s (s:%s, l:%s)", t.ID, t.Name, t.Seeders, t.Leechers)
}

// used for sorting
type torrentsList []Torrent

// bySeeder implements sort.Interface by providing Less and using the Len and
// Swap methods of the embedded Torrents value.
type bySeeder struct {
	torrentsList
}

// Less implements the sort.Interface
func (s bySeeder) Less(i, j int) bool {
	seederI, _ := strconv.Atoi(s.torrentsList[i].Seeders)
	seederJ, _ := strconv.Atoi(s.torrentsList[j].Seeders)
	return seederI < seederJ
}

// Len implements the sort.Interface
func (t torrentsList) Len() int {
	return len(t)
}

// Swap implements the sort.Interface
func (t torrentsList) Swap(i, j int) { t[i], t[j] = t[j], t[i] }

// T411 search API is quite strange to use. see https://api.t411.li/
// they use 'terms' to allow search by category.
// In this case we are only interested in category Season and Episode number.
// Season and Episode number also have specific ID. init method creates the mapping
var (
	catSeasonID   = 45
	catEpisodeID  = 46
	catLanguageID = 51
	seasonNbrID   = map[int]int{}
	episodeNbrID  = map[int]int{}
	languageMap   = map[string]int{
		"english":    1209,
		"french":     1210,
		"mute":       1211,
		"multi-fr":   1212,
		"multi-qb":   1213,
		"quebecker ": 1214,
		"vfstfr":     1215,
		"vostfr":     1216,
	}
)

func init() {
	for i := 0; i < 30; i++ {
		seasonNbrID[i+1] = 968 + i
	}
	for i := 0; i < 60; i++ {
		episodeNbrID[i+1] = 937 + i
	}
}

func addEpisode(v url.Values, episode int) {
	if episode > 0 {
		v.Add(fmt.Sprintf("term[%d][]", catEpisodeID), fmt.Sprintf("%d", episodeNbrID[episode]))
	}
}

func addSeason(v url.Values, season int) {
	if season > 0 {
		v.Add(fmt.Sprintf("term[%d][]", catSeasonID), fmt.Sprintf("%d", seasonNbrID[season]))
	}
}

func addLanguage(v url.Values, language string) {
	if ID, ok := languageMap[language]; ok {
		v.Add(fmt.Sprintf("term[%d][]", catLanguageID), fmt.Sprintf("%d", ID))
	}
}

func addOffset(v url.Values, offset int) {
	if offset >= 0 {
		v.Add("offset", fmt.Sprintf("%d", offset))
	} else {
		v.Add("offset", fmt.Sprintf("%d", 0))
	}
}

func addLimit(v url.Values, limit int) {
	if limit > 0 {
		v.Add("limit", fmt.Sprintf("%d", limit))
	} else {
		v.Add("limit", fmt.Sprintf("%d", 10))
	}
}

// URL returns the url of the search request
func makeURL(title string, season, episode int, language string, offset, limit int) (string, *url.URL, error) {
	usedAPI := "/torrents/search/"
	u, err := url.Parse(fmt.Sprintf("%s%s%s", t411BaseURL, usedAPI, title))
	if err != nil {
		return usedAPI, nil, err
	}
	v := u.Query()
	addSeason(v, season)
	addEpisode(v, episode)
	addLanguage(v, language)
	// there is a bug in the t411 api: if we do request with limit and/or offset parameters
	// corresponding response fields 'limit' and/or 'offset will be of type string. If not,
	// they will be of type int. So we always make requests with those parameters and use the
	// default values of limit/offset parameters if need be.
	addOffset(v, offset)
	addLimit(v, limit)
	u.RawQuery = v.Encode()
	return usedAPI, u, nil
}

// SearchTorrentsByTerms searches a torrent using terms and return a list of torrents
// with a maximum of 10 torrents.
func (t *T411) SearchTorrentsByTerms(title string, season, episode int, language string, offset, limit int) (*Torrents, error) {
	usedAPI, u, err := makeURL(title, season, episode, language, offset, limit)
	if err != nil {
		return nil, err
	}
	resp, err := t.doGet(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	torrents := &Torrents{}
	err = t.decode(torrents, resp, usedAPI, u.RawQuery)
	if err != nil {
		return nil, err
	}
	return torrents, nil
}

// SortBySeeders sorts the given torrents by seeders.
func (*T411) SortBySeeders(torrents []Torrent) {
	sort.Sort(bySeeder{torrents})
}

func (t *T411) download(ID string) (io.ReadCloser, error) {
	u, err := url.Parse(fmt.Sprintf("%s/torrents/download/%s", t.baseURL, ID))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := t.do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, err
}

// DownloadTorrentByID downloads the torrent of id 'id' into a temporary
// filename begining with 'prefix' and returns the complete temporary filename
// on success.
func (t *T411) DownloadTorrentByID(id, prefix string) (string, error) {
	r, err := t.download(id)
	if err != nil {
		return "", err
	}
	defer r.Close()

	tmpfile, err := ioutil.TempFile("", prefix)
	if err != nil {
		return "", err
	}
	defer tmpfile.Close()

	if _, err = io.Copy(tmpfile, r); err != nil {
		return "", err
	}

	return tmpfile.Name(), nil
}

// DownloadTorrentByTerms searches the torrent corresponding to the title,
// season, episode and language, downloads the one with the most seeders
// and return the location of the file located in a temporary folder.
// Note: the search is done with an offset of 0 and a limit of 10 results per search.
// Note: the language parameter must be one of those values: "english", "french",
// "mute", "multi-fr", "multi-qb", "quebecker ", "vfstfr", "vostfr".
func (t *T411) DownloadTorrentByTerms(title string, season, episode int, language string) (string, error) {
	torrents, err := t.SearchTorrentsByTerms(title, season, episode, language, 0, 0)
	if err != nil {
		return "", err
	}

	if len(torrents.Torrents) < 1 {
		return "", fmt.Errorf("torrent %s S%02dE%02d not found", title, season, episode)
	}

	t.SortBySeeders(torrents.Torrents)
	return t.DownloadTorrentByID(torrents.Torrents[len(torrents.Torrents)-1].ID, fmt.Sprintf("%sS%02dE%02d", title, season, episode))
}
