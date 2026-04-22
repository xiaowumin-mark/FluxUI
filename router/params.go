package router

// Params 保存路由参数（路径参数 + 查询参数）。
type Params struct {
	pathParams  map[string]string
	queryParams map[string]string
}

// Get 返回路径参数值（如 ":id"）。
func (p *Params) Get(name string) string {
	if p == nil || p.pathParams == nil {
		return ""
	}
	return p.pathParams[name]
}

// Query 返回查询参数值（如 "?tab=posts"）。
func (p *Params) Query(name string) string {
	if p == nil || p.queryParams == nil {
		return ""
	}
	return p.queryParams[name]
}

// Path 返回路径参数值（Get 的别名）。
func (p *Params) Path(name string) string {
	return p.Get(name)
}

// HasParam 判断是否存在指定路径参数。
func (p *Params) HasParam(name string) bool {
	if p == nil || p.pathParams == nil {
		return false
	}
	_, ok := p.pathParams[name]
	return ok
}

// HasQuery 判断是否存在指定查询参数。
func (p *Params) HasQuery(name string) bool {
	if p == nil || p.queryParams == nil {
		return false
	}
	_, ok := p.queryParams[name]
	return ok
}

// AllPathParams 返回所有路径参数的副本。
func (p *Params) AllPathParams() map[string]string {
	if p == nil || p.pathParams == nil {
		return nil
	}
	out := make(map[string]string, len(p.pathParams))
	for k, v := range p.pathParams {
		out[k] = v
	}
	return out
}

// AllQueryParams 返回所有查询参数的副本。
func (p *Params) AllQueryParams() map[string]string {
	if p == nil || p.queryParams == nil {
		return nil
	}
	out := make(map[string]string, len(p.queryParams))
	for k, v := range p.queryParams {
		out[k] = v
	}
	return out
}
