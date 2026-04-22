package router

import "testing"

func TestMatchPathStatic(t *testing.T) {
	result := matchPath("/home", "/home")
	if !result.matched {
		t.Fatal("expected /home to match /home")
	}
}

func TestMatchPathStaticMismatch(t *testing.T) {
	result := matchPath("/home", "/about")
	if result.matched {
		t.Fatal("expected /home not to match /about")
	}
}

func TestMatchPathParam(t *testing.T) {
	result := matchPath("/users/:id", "/users/42")
	if !result.matched {
		t.Fatal("expected match")
	}
	if result.params["id"] != "42" {
		t.Fatalf("expected id=42, got %s", result.params["id"])
	}
}

func TestMatchPathMultipleParams(t *testing.T) {
	result := matchPath("/users/:uid/posts/:pid", "/users/10/posts/99")
	if !result.matched {
		t.Fatal("expected match")
	}
	if result.params["uid"] != "10" {
		t.Fatalf("expected uid=10, got %s", result.params["uid"])
	}
	if result.params["pid"] != "99" {
		t.Fatalf("expected pid=99, got %s", result.params["pid"])
	}
}

func TestMatchPathTooShort(t *testing.T) {
	result := matchPath("/users/:id/posts", "/users/42")
	if result.matched {
		t.Fatal("expected no match for shorter path")
	}
}

func TestMatchPathTooLong(t *testing.T) {
	result := matchPath("/users", "/users/42")
	if result.matched {
		t.Fatal("expected no match for longer path")
	}
}

func TestMatchPathRoot(t *testing.T) {
	result := matchPath("/", "/")
	if !result.matched {
		t.Fatal("expected root match")
	}
}

func TestMatchPathWildcard(t *testing.T) {
	result := matchPath("/*", "/anything/here")
	if !result.matched {
		t.Fatal("expected wildcard match")
	}
}

func TestExtractQueryParams(t *testing.T) {
	path, query := extractQueryParams("/users/42?tab=posts&sort=asc")
	if path != "/users/42" {
		t.Fatalf("expected path=/users/42, got %s", path)
	}
	if query["tab"] != "posts" {
		t.Fatalf("expected tab=posts, got %s", query["tab"])
	}
	if query["sort"] != "asc" {
		t.Fatalf("expected sort=asc, got %s", query["sort"])
	}
}

func TestExtractQueryParamsNone(t *testing.T) {
	path, query := extractQueryParams("/users/42")
	if path != "/users/42" {
		t.Fatalf("expected path=/users/42, got %s", path)
	}
	if query != nil {
		t.Fatal("expected nil query")
	}
}

func TestParamsGet(t *testing.T) {
	p := &Params{
		pathParams:  map[string]string{"id": "42"},
		queryParams: map[string]string{"tab": "posts"},
	}
	if p.Get("id") != "42" {
		t.Fatal("expected id=42")
	}
	if p.Query("tab") != "posts" {
		t.Fatal("expected tab=posts")
	}
	if p.Get("missing") != "" {
		t.Fatal("expected empty for missing param")
	}
	if p.Query("missing") != "" {
		t.Fatal("expected empty for missing query")
	}
}

func TestParamsNil(t *testing.T) {
	var p *Params
	if p.Get("id") != "" {
		t.Fatal("expected empty from nil params")
	}
	if p.Query("tab") != "" {
		t.Fatal("expected empty from nil params")
	}
}

func TestSplitPath(t *testing.T) {
	cases := []struct {
		input string
		want  []string
	}{
		{"/", nil},
		{"/home", []string{"home"}},
		{"/users/42/posts", []string{"users", "42", "posts"}},
		{"users/42", []string{"users", "42"}},
		{"/a//b/", []string{"a", "b"}},
	}

	for _, tc := range cases {
		got := splitPath(tc.input)
		if len(got) == 0 && len(tc.want) == 0 {
			continue
		}
		if len(got) != len(tc.want) {
			t.Fatalf("splitPath(%q): want %v, got %v", tc.input, tc.want, got)
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Fatalf("splitPath(%q)[%d]: want %q, got %q", tc.input, i, tc.want[i], got[i])
			}
		}
	}
}

func TestReverseTransition(t *testing.T) {
	if reverseTransition(TransitionSlideLeft) != TransitionSlideRight {
		t.Fatal("SlideLeft should reverse to SlideRight")
	}
	if reverseTransition(TransitionSlideRight) != TransitionSlideLeft {
		t.Fatal("SlideRight should reverse to SlideLeft")
	}
	if reverseTransition(TransitionFade) != TransitionFade {
		t.Fatal("Fade should reverse to Fade")
	}
	if reverseTransition(TransitionNone) != TransitionNone {
		t.Fatal("None should reverse to None")
	}
}
