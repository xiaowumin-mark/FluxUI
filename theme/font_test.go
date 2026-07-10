package theme

import "testing"

func TestFontFaceDataReturnsCopy(t *testing.T) {
	source := &fontSource{data: []byte{1, 2, 3, 4}}
	first := FontFace{source: source}
	second := FontFace{source: source}

	data := first.Data()
	data[0] = 99

	if source.data[0] != 1 {
		t.Fatalf("FontFace.Data mutated shared source: %#v", source.data)
	}
	other := second.Data()
	if other[0] != 1 {
		t.Fatalf("FontFace.Data exposed mutation through another face: %#v", other)
	}
}
