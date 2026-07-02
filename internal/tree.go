package internal

// TreePath 返回当前组件树路径，便于调试和扩展。
func (c *Context) TreePath() string {
	if c == nil {
		return ""
	}
	if c.runtime != nil {
		return c.runtime.DebugPath(c.pathID)
	}
	if c.debugPath == "" {
		return "root"
	}
	return c.debugPath
}
