package search

import (
	"testing"
)

func TestEngineRotatorAntiRepeat(t *testing.T) {
	webEnginePool := []string{
		"google", "bing", "duckduckgo", "brave", "wikipedia", "yahoo",
		"baidu", "naver", "startpage", "wiby", "qwant", "mojeek",
		"sogou", "so360", "seznam", "presearch", "marginalia", "webcrawler",
		"excite", "metacrawler", "ecosia", "swisscows", "yandex", "yep",
		"blackle", "dogpile", "aol", "mailru",
	}

	rotator := NewEngineRotator(webEnginePool)

	var prevBatch []string
	allSeen := make(map[string]int)

	for q := 1; q <= 20; q++ {
		batch := rotator.NextBatch(5)
		if len(batch) != 5 {
			t.Fatalf("Query #%d: expected 5 engines, got %d", q, len(batch))
		}

		// Check uniqueness within the batch
		seenInBatch := make(map[string]bool)
		for _, eng := range batch {
			if seenInBatch[eng] {
				t.Fatalf("Query #%d: duplicate engine '%s' within same batch", q, eng)
			}
			seenInBatch[eng] = true
			allSeen[eng]++
		}

		// Check anti-repeat from previous batch
		if len(prevBatch) > 0 {
			prevSet := make(map[string]bool)
			for _, p := range prevBatch {
				prevSet[p] = true
			}
			for _, eng := range batch {
				if prevSet[eng] {
					t.Fatalf("Query #%d: engine '%s' was repeated from previous query batch %v", q, eng, prevBatch)
				}
			}
		}

		t.Logf("Query #%d picked: %v", q, batch)
		prevBatch = batch
	}

	t.Logf("All 28 engines coverage across 20 queries: %d distinct engines picked", len(allSeen))
	if len(allSeen) < 25 {
		t.Fatalf("Expected wide coverage of engine pool, only got %d engines", len(allSeen))
	}
}
