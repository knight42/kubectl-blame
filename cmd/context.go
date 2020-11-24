package cmd

type Context struct {
	Level   int
	NewLine bool
	Node    *Node
}

func (c Context) WithLevel(lvl int) Context {
	c.Level = lvl
	return c
}

func (c Context) WithNewLine(newLine bool) Context {
	c.NewLine = newLine
	return c
}

func (c Context) WithNode(node *Node) Context {
	c.Node = node
	return c
}
