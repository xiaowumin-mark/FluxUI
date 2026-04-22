package router

import "strings"

// matchResult 保存路径匹配的结果。
type matchResult struct {
	matched bool
	params  map[string]string
}

// matchPath 将 URL 路径与路由模式匹配。
// 模式段以 ":" 开头为动态参数（如 "/users/:id"）。
// 模式段 "*" 为通配符，匹配剩余所有路径段。
func matchPath(pattern, path string) matchResult {
	patternSegments := splitPath(pattern)
	pathSegments := splitPath(path)

	params := map[string]string{}

	for i, seg := range patternSegments {
		// 通配符匹配剩余
		if seg == "*" {
			return matchResult{matched: true, params: params}
		}

		if i >= len(pathSegments) {
			return matchResult{matched: false}
		}

		if strings.HasPrefix(seg, ":") {
			paramName := seg[1:]
			params[paramName] = pathSegments[i]
		} else if seg != pathSegments[i] {
			return matchResult{matched: false}
		}
	}

	if len(patternSegments) != len(pathSegments) {
		return matchResult{matched: false}
	}

	return matchResult{matched: true, params: params}
}

// splitPath 将路径按 "/" 分割，忽略空段。
func splitPath(path string) []string {
	parts := strings.Split(path, "/")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// extractQueryParams 从路径中解析出纯路径和查询参数。
func extractQueryParams(fullPath string) (path string, query map[string]string) {
	idx := strings.IndexByte(fullPath, '?')
	if idx < 0 {
		return fullPath, nil
	}

	path = fullPath[:idx]
	queryStr := fullPath[idx+1:]
	query = map[string]string{}

	pairs := strings.Split(queryStr, "&")
	for _, pair := range pairs {
		if pair == "" {
			continue
		}
		eqIdx := strings.IndexByte(pair, '=')
		if eqIdx < 0 {
			query[pair] = ""
		} else {
			query[pair[:eqIdx]] = pair[eqIdx+1:]
		}
	}

	return path, query
}
