package generator

import "testing"

func isAllowedChar(b byte) bool {
	return (b >= 'a' && b <= 'z') ||
		(b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9') ||
		b == '_'
}

func TestGenerateShortUrl_LengthIsTen(t *testing.T) {
	short, err := GenerateShortUrl()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(short) != 10 {
		t.Fatalf("expected length 10, got %d (%q)", len(short), short)
	}
}

func TestGenerateShortUrl_OnlyAllowedCharacters(t *testing.T) {

	for i := 0; i < 1000; i++ {
		short, err := GenerateShortUrl()
		if err != nil {
			t.Fatalf("unexpected error on iteration %d: %v", i, err)
		}

		for j := 0; j < len(short); j++ {
			if !isAllowedChar(short[j]) {
				t.Fatalf("invalid char %q at iteration %d in %q", short[j], i, short)
			}
		}
	}
}
