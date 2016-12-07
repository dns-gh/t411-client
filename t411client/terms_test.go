package t411client

import (
	. "gopkg.in/check.v1"
)

func checkTermsTree(c *C, termsTree map[string]ByTermID) {
	for k1, v1 := range termsTree {
		c.Assert(k1, Not(HasLen), 0)
		c.Assert(v1, Not(HasLen), 0)
		for k2, v2 := range v1 {
			c.Assert(k2, Not(HasLen), 0)
			c.Assert(v2.Mode, Not(HasLen), 0)
			c.Assert(v2.Type, Not(HasLen), 0)
			c.Assert(v2.Terms, Not(HasLen), 0)
			for k3, v3 := range v2.Terms {
				c.Assert(k3, Not(HasLen), 0)
				c.Assert(v3, Not(HasLen), 0)
			}
		}
	}
}

func (s *MySuite) TestTermsTree(c *C) {
	t411, _, _ := createT411Client(c)
	termsTree, err := t411.TermsTree()
	c.Assert(err, IsNil)
	c.Assert(termsTree.ByCategoryID, Not(HasLen), 0)
	checkTermsTree(c, termsTree.ByCategoryID)
}
