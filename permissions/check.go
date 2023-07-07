package permissions

type Checker struct {
	permission Permission
	res        bool
	finished   bool
}

func (c Checker) Res() bool {
	return c.res
}

func (c *Checker) And(str string) *Checker {
	if c.finished {
		return c
	}
	if c.permission.Has(str) {
		c.res = true
	} else {
		c.res = false
		c.finished = true
	}
	return c
}

func (c *Checker) Or(str string) *Checker {
	c.finished = false
	if c.permission.Has(str) {
		c.res = true
	} else {
		c.finished = false
	}
	return c
}

func (p Permission) Check(str string) *Checker {
	return &Checker{
		permission: p,
		res:        p.Has(str),
	}
}
