package entities

type ContentTest struct {
	BaseEntity
}

func (c *ContentTest) test() {
	c.Save()
}
