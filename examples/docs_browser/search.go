package main

import "strings"

func filterDocs(docs []widgetDoc, keyword string) []widgetDoc {
	terms := docsSearchTerms(keyword)
	if len(terms) == 0 {
		return docs
	}

	out := make([]widgetDoc, 0, len(docs))
	for i := range docs {
		doc := docs[i]
		haystack := docsSearchHaystack(doc)
		if docsMatchesSearchTerms(haystack, terms) {
			out = append(out, doc)
		}
	}
	return out
}

func docsSearchTerms(keyword string) []string {
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(keyword)))
	if len(fields) == 0 {
		return nil
	}
	terms := make([]string, 0, len(fields))
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field != "" {
			terms = append(terms, field)
		}
	}
	return terms
}

func docsSearchHaystack(doc widgetDoc) string {
	return strings.ToLower(strings.Join([]string{
		doc.Meta.ID,
		doc.Meta.Title,
		doc.Meta.Category,
		doc.Meta.Summary,
		strings.Join(doc.Meta.APIs, " "),
		doc.Content,
	}, " "))
}

func docsMatchesSearchTerms(haystack string, terms []string) bool {
	haystack = strings.ToLower(haystack)
	for _, term := range terms {
		if !strings.Contains(haystack, term) {
			return false
		}
	}
	return true
}
