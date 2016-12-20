package t411client

import (
	"net/url"
)

// Term represents the term data
type Term struct {
	Type  string            `json:"type"`
	Mode  string            `json:"mode"`
	Terms map[string]string `json:"terms"`
}

// ByTermID represents a list of terms mapped by their id
type ByTermID map[string]Term

// TermsTree represents the terms tree
type TermsTree struct {
	ByCategoryID map[string]ByTermID
}

// TermsTree gets the terms tree
func (t *T411) TermsTree() (*TermsTree, error) {
	usedAPI := "/terms/tree"
	u, err := url.Parse(t.baseURL + usedAPI)
	if err != nil {
		return nil, errURLParsing
	}

	resp, err := t.do("GET", u, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	termsTree := &TermsTree{}
	err = t.decode(&termsTree.ByCategoryID, resp, usedAPI, u.RawQuery)
	if err != nil {
		return nil, err
	}
	return termsTree, nil
}
