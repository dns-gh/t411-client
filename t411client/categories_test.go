package t411client

import (
	"fmt"

	. "gopkg.in/check.v1"
)

func checkCategories(c *C, categories map[string]Category) {
	for k, v := range categories {
		if len(v.ID) == 0 {
			fmt.Println("one category with no id found")
			continue
		}
		c.Assert(k, Not(HasLen), 0)
		c.Assert(v.ID, Not(HasLen), 0)
		c.Assert(v.Pid, Not(HasLen), 0)
		c.Assert(v.Name, Not(HasLen), 0)
		checkCategories(c, v.Cats)
	}
}

func (s *MySuite) TestCategoriesTree(c *C) {
	t411, _, _ := createT411Client(c)
	categories, err := t411.CategoriesTree()
	c.Assert(err, IsNil)
	c.Assert(categories.Categories, Not(HasLen), 0)
	checkCategories(c, categories.Categories)
}
