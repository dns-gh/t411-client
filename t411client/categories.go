package t411client

import (
	"net/url"
)

// Category represents the category data
type Category struct {
	ID   string              `json:"id"`
	Pid  string              `json:"pid"`
	Name string              `json:"name"`
	Cats map[string]Category `json:"cats"`
}

// Categories represents the categories data
type Categories struct {
	Categories map[string]Category
}

// CategoriesTree gets the categories tree
func (t *T411) CategoriesTree() (*Categories, error) {
	usedAPI := "/categories/tree"
	u, err := url.Parse(t.baseURL + usedAPI)
	if err != nil {
		return nil, errURLParsing
	}

	resp, err := t.doGet(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	categories := &Categories{}
	err = t.decode(&categories.Categories, resp, usedAPI, u.RawQuery)
	if err != nil {
		return nil, err
	}
	return categories, nil
}
