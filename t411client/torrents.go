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
	errEOF             = errors.New("no more torrents to find")
	errTorrentNotFound = &errAPI{
		Code: 301,
		Text: "Torrent not found",
	}
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
	catSeasonID       = 45
	catEpisodeID      = 46
	catLanguageID     = 51
	catQualityID      = 7
	episodeNbrIDStart = 936
	seasonNbrIDStart  = 968
	seasonNbrID       = map[int]int{}
	episodeNbrID      = map[int]int{}
	// LanguageMap is a map giving you the keys corresponding
	// to every available language filter
	LanguageMap = map[string]int{
		"english":    1209,
		"french":     1210,
		"mute":       1211,
		"multi-fr":   1212,
		"multi-qb":   1213,
		"quebecker ": 1214,
		"vfstfr":     1215,
		"vostfr":     1216,
		"voasta":     1217,
	}
	// QualityMap is a map giving you the keys corresponding
	// to every available quality filter
	QualityMap = map[string]int{
		"BDrip/BRrip [Rip SD (non HD) depuis Bluray ou HDrip": 8,
		"Bluray 4K [Full ou Remux]":                           1171,
		"Bluray [Full]":                                       17,
		"Bluray [Remux]":                                      1220,
		"DVD-R 5 [DVD < 4.37GB]":                              13,
		"DVD-R 9 [DVD > 4.37GB]":                              14,
		"DVDrip [Rip depuis DVD-R]":                           10,
		"HDrip 1080 [Rip HD depuis Bluray]":                   16,
		"HDrip 4k [Rip HD 4k depuis source 4k]":               1219,
		"HDrip 720 [Rip HD depuis Bluray]":                    15,
		"TVrip [Rip SD (non HD) depuis Source Tv HD/SD]":      11,
		"TVripHD 1080 [Rip HD depuis Source Tv HD]":           1162,
		"TvripHD 4k [Rip HD 4k depuis Source Tv 4k]":          1235,
		"TVripHD 720 [Rip HD depuis Source Tv HD]":            12,
		"VCD/SVCD/VHSrip":                                     18,
		"Web-Dl":                                              1233,
		"Web-Dl 1080":                                         1174,
		"Web-Dl 4K":                                           1182,
		"Web-Dl 720":                                          1175,
		"WEBrip":                                              19,
	}
)

func init() {
	for i := 0; i < 30; i++ {
		seasonNbrID[i+1] = seasonNbrIDStart + i
	}
	seasonNbrID[0] = 998
	for i := 0; i < 61; i++ {
		episodeNbrID[i] = episodeNbrIDStart + i
	}
}

func addEpisode(v url.Values, episode int) {
	if episode >= 0 {
		v.Add(fmt.Sprintf("term[%d][]", catEpisodeID), fmt.Sprintf("%d", episodeNbrID[episode]))
	}
}

func addSeason(v url.Values, season int) {
	if season >= 0 {
		v.Add(fmt.Sprintf("term[%d][]", catSeasonID), fmt.Sprintf("%d", seasonNbrID[season]))
	}
}

func addLanguage(v url.Values, language string) {
	if ID, ok := LanguageMap[language]; ok {
		v.Add(fmt.Sprintf("term[%d][]", catLanguageID), fmt.Sprintf("%d", ID))
	}
}

func addQuality(v url.Values, quality string) {
	if ID, ok := QualityMap[quality]; ok {
		v.Add(fmt.Sprintf("term[%d][]", catQualityID), fmt.Sprintf("%d", ID))
	}
}

func addOffset(v url.Values, offset int) {
	if offset >= 0 {
		v.Add("offset", fmt.Sprintf("%d", offset))
	}
}

func addLimit(v url.Values, limit int) {
	if limit > 0 {
		v.Add("limit", fmt.Sprintf("%d", limit))
	}
}

// URL returns the url of the search request
func makeURL(title string, season, episode int, language, quality string, offset, limit int) (string, *url.URL, error) {
	usedAPI := "/torrents/search/"
	u, err := url.Parse(fmt.Sprintf("%s%s%s", t411BaseURL, usedAPI, title))
	if err != nil {
		return usedAPI, nil, err
	}
	v := u.Query()
	addSeason(v, season)
	addEpisode(v, episode)
	addLanguage(v, language)
	addQuality(v, quality)
	addOffset(v, offset)
	addLimit(v, limit)
	u.RawQuery = v.Encode()
	return usedAPI, u, nil
}

// SearchTorrentsByTerms searches a torrent using terms and return a list of torrents
// with a maximum of 10 torrents by default.
// The 'title' parameter is mandatory of course. All the others are optionals.
// For 'season' and 'episode', a value of 0 means respectively "complete/integral tv show" and "complete season",
// Note that for the complete tv show, both 'season' and 'episode' must be set to 0.
// 'season' available range is from 0 to 30 and 'episode' range is from 0 to 60.
// The 'language' parameter must be one the values of the LanguageMap variable.
// The 'quality' parameter must be one the values of the QualityMap variable.
func (t *T411) SearchTorrentsByTerms(title string, season, episode int, language, quality string, offset, limit int) (*Torrents, error) {
	usedAPI, u, err := makeURL(title, season, episode, language, quality, offset, limit)
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

// SearchAllTorrentByTerms does the same as SearchTorrentByTerms but get all the possible torrents
// for the specific search in a single torrent slice.
func (t *T411) SearchAllTorrentByTerms(title string, season, episode int, language, quality string) (*Torrents, error) {
	torrents, err := t.SearchTorrentsByTerms(title, season, episode, language, quality, 0, 100)
	if err != nil {
		return nil, err
	}
	total, err := strconv.Atoi(torrents.Total)
	if err != nil {
		return nil, err
	}
	if len(torrents.Torrents) == total {
		return torrents, nil
	}

	torrents, err = t.SearchTorrentsByTerms(title, season, episode, language, quality, 0, total)
	if err != nil {
		return nil, err
	}
	return torrents, nil
}

// TorrentDetails represents the torrent detail data.
type TorrentDetails struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Category      string            `json:"category"`
	Categoryname  string            `json:"categoryname"`
	Categoryimage string            `json:"categoryimage"`
	Rewritename   string            `json:"rewritename"`
	Owner         string            `json:"owner"`
	Username      string            `json:"username"`
	Privacy       string            `json:"privacy"`
	Description   string            `json:"description"`
	Terms         map[string]string `json:"terms"`
}

// TorrentsDetails returns the details of a torrent from a torrent 'id'.
func (t *T411) TorrentsDetails(id string) (*TorrentDetails, error) {
	usedAPI := "/torrents/details/"
	u, err := url.Parse(fmt.Sprintf("%s%s%s", t411BaseURL, usedAPI, id))
	if err != nil {
		return nil, err
	}
	resp, err := t.doGet(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	details := &TorrentDetails{}
	err = t.decode(details, resp, usedAPI, u.RawQuery)
	if err != nil {
		return nil, err
	}
	return details, nil
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
// Note: the search is done with an offset of 0 and a limit of 10 results per search by default.
// Note: the 'language' parameter must be one of the values of LanguageMap variable.
// Note: the 'quality' parameter must be one of the values of QualityMap variable.
func (t *T411) DownloadTorrentByTerms(title string, season, episode int, language, quality string) (string, error) {
	torrents, err := t.SearchTorrentsByTerms(title, season, episode, language, quality, 0, 0)
	if err != nil {
		return "", err
	}

	t.SortBySeeders(torrents.Torrents)
	return t.DownloadTorrentByID(torrents.Torrents[len(torrents.Torrents)-1].ID, fmt.Sprintf("%sS%02dE%02d", title, season, episode))
}
