package internal

// TreePath 返回当前组件树路径，便于调试和扩展。
func (c *Context) TreePath() string {
	return c.path
}
