package t411client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
)

var (
	errEOF = errors.New("no more torrents to find")
)

// Torrent represent a torrent as return by the t411 API
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

// Torrents represents an array of torrents. Mainly used for sorting.
type Torrents []Torrent

func (t *Torrent) String() string {
	return fmt.Sprintf("%s - %s (%s)", t.ID, t.Name, t.Seeders)
}

// BySeeder implements sort.Interface by providing Less and using the Len and
// Swap methods of the embedded Torrents value.
type BySeeder struct {
	Torrents
}

// Less implements the sort.Interface
func (s BySeeder) Less(i, j int) bool {
	seederI, _ := strconv.Atoi(s.Torrents[i].Seeders)
	seederJ, _ := strconv.Atoi(s.Torrents[j].Seeders)
	return seederI < seederJ
}

// Len implements the sort.Interface
func (t Torrents) Len() int {
	return len(t)
}

// Swap implements the sort.Interface
func (t Torrents) Swap(i, j int) { t[i], t[j] = t[j], t[i] }

// T411 search API is quite strange to use. see https://api.t411.li/
// they use 'terms' to allow search by category.
// In this case we are only interested in category Season and episode number.
// Season and episode number also have specific ID. init method creates the mapping

var (
	catSeasonID   = 45
	catEpisodeID  = 46
	catLanguageID = 51
	seasonNbrID   = map[int]int{}
	episodeNbrID  = map[int]int{}
	languageMap   = map[string]int{
		"anglais":   1209,
		"français":  1210,
		"muet":      1211,
		"multi-fr":  1212,
		"multi-qb":  1213,
		"québécois": 1214,
		"vfstfr":    1215,
		"vostfr":    1216,
		"voasta":    1217,
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

type searchReq struct {
	Title    string
	Season   int
	Episode  int
	Language string
}

// URL returns the url of the search request
func (r searchReq) URL() string {
	u, err := url.Parse(fmt.Sprintf("%s/torrents/search/%s", t411BaseURL, r.Title))
	if err != nil {
		log.Fatalf("Error during construction of t411 search URL: %v", err)
	}
	q := u.Query()
	if r.Season > 0 {
		q.Add(fmt.Sprintf("term[%d][]", catSeasonID), fmt.Sprintf("%d", seasonNbrID[r.Season]))
	}
	if r.Episode > 0 {
		q.Add(fmt.Sprintf("term[%d][]", catEpisodeID), fmt.Sprintf("%d", episodeNbrID[r.Episode]))
	}
	if ID, ok := languageMap[r.Language]; ok {
		q.Add(fmt.Sprintf("term[%d][]", catLanguageID), fmt.Sprintf("%d", ID))

	}
	u.RawQuery = q.Encode()

	return u.String()
}

func (t *T411) search(searchReq searchReq) ([]Torrent, error) {

	req, err := http.NewRequest("GET", searchReq.URL(), nil)
	if err != nil {
		log.Printf("Error creating request to %s: %v", searchReq.URL(), err)
		return nil, err
	}

	resp, err := t.do(req)
	if err != nil {
		log.Printf("Error executing request to %s: %v", searchReq.URL(), err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatal("bad status code", resp.StatusCode)
	}

	data := struct {
		Torrents []Torrent `json:"torrents"`
	}{}

	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.Torrents, nil
}

// DownloadTorrent search the torrent corresponding to the title,
// season and episode number, download it and return the location of the file.
func (t *T411) DownloadTorrent(title string, season, episode int, language string) (string, error) {
	req := searchReq{
		Title:    title,
		Season:   season,
		Episode:  episode,
		Language: language,
	}

	torrents, err := t.search(req)
	if err != nil {
		log.Printf("Error search for torrent: %v", err.Error())
		return "", err
	}

	if len(torrents) < 1 {
		return "", fmt.Errorf("torrent not found, %sS%02dE%02d", title, season, episode)
	}

	sort.Sort(BySeeder{torrents})

	r, err := t.download(torrents[len(torrents)-1].ID)
	if err != nil {
		return "", err
	}
	defer r.Close()

	tmpfile, err := ioutil.TempFile("", fmt.Sprintf("%sS%02dE%02d", title, season, episode))
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer tmpfile.Close()

	if _, err = io.Copy(tmpfile, r); err != nil {
		log.Println(err)
		return "", err
	}

	return tmpfile.Name(), nil
}

func (t *T411) download(ID string) (io.ReadCloser, error) {

	u, err := url.Parse(fmt.Sprintf("%s/torrents/download/%s", t.baseURL, ID))
	if err != nil {
		log.Println("Error parsing url: ", err)
		return nil, err
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Println("Error creating downlaod request: ", err)
		return nil, err
	}

	resp, err := t.do(req)
	if err != nil {
		log.Println("Error executing download request: ", err)
		return nil, err
	}

	return resp.Body, err
}

//
// func chooseTorrent(torrents []Torrent) (string, error) {
// 	if len(torrents) < 1 {
// 		return "", fmt.Errorf("no torrents to choose from")
// 	}
//
// 	sort.Sort(sort.Reverse(BySeeder{torrents}))
//
// 	for i, torrent := range torrents {
// 		fmt.Printf("%d %s\n", i+1, torrent.String())
// 	}
// 	fmt.Printf("Which torrent do you want ? (1-%d) : \n", len(torrents))
// 	var index int
// 	fmt.Scanf("%d", &index)
//
// 	return torrents[index-1].ID, nil
// }
