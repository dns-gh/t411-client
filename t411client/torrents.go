package t411client

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	errEOF = errors.New("no more torrents to find")
	// it seems like a bug in the t411 API where two identicals errors
	// have different error codes. The one to remove would be the err301...
	err301TorrentNotFound = &errAPI{
		Code: 301,
		Text: "Torrent not found",
	}
	//ErrTorrentNotFound represents the 1301 error code 'torrent not found'.
	ErrTorrentNotFound = &errAPI{
		Code: 1301,
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

func (t *Torrent) checkTorrentName(title string) bool {
	return strings.Contains(strings.ToLower(t.Name), strings.ToLower(title))
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
	catSeasonID          = 45
	catEpisodeID         = 46
	catLanguageID        = 51
	catQualityID         = 7
	episodeNbrIDStart    = 936
	episodeNbrIDMiddle   = 1088
	seasonNbrIDStart     = 968
	seasonCompleteSeries = 998
	seasonNbrID          = map[int]int{}
	episodeNbrID         = map[int]int{}
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
	seasonNbrID[0] = seasonCompleteSeries
	// see https://github.com/dns-gh/t411-client/issues/3 for more explanation
	for i := 0; i < 31; i++ {
		if i == 9 {
			episodeNbrIDStart++
		}
		episodeNbrID[i] = episodeNbrIDStart + i
		if i != 30 {
			episodeNbrID[31+i] = episodeNbrIDMiddle + i
		}
	}
	// switch 16 <-> 17 episode values. See:
	// <option value="954">Episode 16</option>
	// <option value="953">Episode 17</option>
	temp := episodeNbrID[16]
	episodeNbrID[16] = episodeNbrID[17]
	episodeNbrID[17] = temp
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
	if ID, ok := LanguageMap[strings.ToLower(language)]; ok {
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
func (t *T411) makeURL(title string, season, episode int, language, quality string, offset, limit int) (string, *url.URL, error) {
	usedAPI := "/torrents/search/"
	u, err := url.Parse(fmt.Sprintf("%s%s%s", t.baseURL, usedAPI, title))
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
	usedAPI, u, err := t.makeURL(title, season, episode, language, quality, offset, limit)
	if err != nil {
		return nil, err
	}
	resp, err := t.do("GET", u, nil)
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
	u, err := url.Parse(fmt.Sprintf("%s%s%s", t.baseURL, usedAPI, id))
	if err != nil {
		return nil, err
	}
	resp, err := t.do("GET", u, nil)
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

func (t *T411) checkRatio(torrent *Torrent) error {
	if t.keepRatio && len(torrent.Size) != 0 {
		torrentSize, err := strconv.ParseFloat(torrent.Size, 64)
		if err != nil {
			return err
		}
		ratio, err := t.GetOwnRatio(torrentSize)
		if err != nil {
			return err
		}
		if ratio < 1 {
			return fmt.Errorf("cannot download to keep ratio > 1")
		}
	}
	return nil
}

func (t *T411) checkVerified(torrent *Torrent) error {
	if t.onlyVerified && torrent.IsVerified == "false" {
		return fmt.Errorf("cannot download non-verified torrent")
	}
	return nil
}

// DownloadTorrent downloads the torrent into a temporary
// folder on success and returns the absolute path to the newly created file.
func (t *T411) DownloadTorrent(torrent *Torrent) (string, error) {
	if err := t.checkRatio(torrent); err != nil {
		return "", fmt.Errorf("cannot download to keep ratio > 1")
	}
	if err := t.checkVerified(torrent); err != nil {
		return "", fmt.Errorf("cannot download non-verified torrent")
	}
	u, err := url.Parse(fmt.Sprintf("%s/torrents/download/%s", t.baseURL, torrent.ID))
	if err != nil {
		return "", err
	}

	resp, err := t.do("GET", u, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bytes, err := decodeErr(resp)
	if err != nil {
		return "", err
	}
	contentDisposition := resp.Header["Content-Disposition"]
	if len(contentDisposition) == 0 {
		return "", fmt.Errorf("'Content-Disposition' header not found in http response")
	}
	split := strings.Split(contentDisposition[0], "\"")
	if len(split) != 3 {
		return "", fmt.Errorf("failed to extract filename from http 'Content-Disposition' header")
	}
	filename := filepath.Join(os.TempDir(), split[1])
	err = ioutil.WriteFile(filename, bytes, 0666)
	if err != nil {
		return "", err
	}
	return filename, nil
}

func (t *T411) filterByPart(torrents []Torrent) ([]Torrent, error) {
	filtered := []Torrent{}
	for _, v := range torrents {
		if !strings.Contains(strings.ToLower(v.Name), ".part.") {
			filtered = append(filtered, v)
		}
	}
	return filtered, nil
}

func (t *T411) filterByDate(torrents []Torrent, date string) ([]Torrent, error) {
	timeConstraint, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, err
	}
	filtered := []Torrent{}
	for _, v := range torrents {
		timeAdded, err := time.Parse("2006-01-02 15:04:05", v.Added)
		if err != nil {
			return nil, err
		}
		diff := timeAdded.Sub(timeConstraint).Hours()
		// 2 weeks close
		if diff >= 0 && diff <= t.maxDelay {
			filtered = append(filtered, v)
		}
	}
	return filtered, nil
}

func (t *T411) filterByName(torrents []Torrent, title string) ([]Torrent, error) {
	title = strings.ToLower(title)
	title = strings.Replace(title, "(", " ", -1)
	title = strings.Replace(title, ")", " ", -1)
	words := strings.Split(title, " ")
	filtered := []Torrent{}
	for _, v := range torrents {
		lowerName := strings.ToLower(v.Name)
		containsAll := true
		for _, v := range words {
			if !strings.Contains(lowerName, v) {
				containsAll = false
				break
			}
		}
		if containsAll {
			filtered = append(filtered, v)
		}
	}
	return filtered, nil
}

// DownloadTorrentByTerms searches the torrent corresponding to the title,
// season, episode and language, downloads the one with the most seeders
// and return the location of the file located in a temporary folder.
// It also filters the torrents by a date in order to get torrents
// with a date close to the provided one, if any.
// Note: the search is done with an offset of 0 and a limit of 10 results per search by default.
// Note: the 'language' parameter must be one of the values of LanguageMap variable.
// Note: the 'quality' parameter must be one of the values of QualityMap variable.
func (t *T411) DownloadTorrentByTerms(title string, season, episode int, language, quality, date string) (string, error) {
	torrents, err := t.SearchTorrentsByTerms(title, season, episode, language, quality, 0, 0)
	if err != nil {
		return "", err
	}
	torrentList := torrents.Torrents
	// filter by name first since pending torrents can have
	// field Name empty so we check that first.
	torrentList, err = t.filterByName(torrentList, title)
	if err != nil {
		return "", err
	}
	if len(date) != 0 {
		torrentList, err = t.filterByDate(torrentList, date)
		if err != nil {
			return "", err
		}
	}
	if season != 0 && episode == 0 {
		torrentList, err = t.filterByPart(torrentList)
		if err != nil {
			return "", err
		}
	}
	if len(torrentList) == 0 {
		return "", ErrTorrentNotFound
	}
	t.SortBySeeders(torrentList)
	torrent := torrentList[len(torrentList)-1]
	return t.DownloadTorrent(&torrent)
}
