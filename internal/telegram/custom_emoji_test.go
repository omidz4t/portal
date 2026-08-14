package telegram

import "testing"

func TestExtractCustomEmojiIDs(t *testing.T) {
	raw := []byte(`{
		"entities": [
			{"type":"custom_emoji","offset":0,"length":2,"custom_emoji_id":"111"},
			{"type":"bold","offset":2,"length":1},
			{"type":"custom_emoji","offset":4,"length":2,"custom_emoji_id":"222"},
			{"type":"custom_emoji","offset":6,"length":2,"custom_emoji_id":"111"}
		],
		"caption_entities": [
			{"type":"custom_emoji","offset":0,"length":2,"custom_emoji_id":"333"}
		]
	}`)
	ids := extractCustomEmojiIDs(raw)
	if len(ids) != 3 {
		t.Fatalf("got %v", ids)
	}
	want := map[string]bool{"111": true, "222": true, "333": true}
	for _, id := range ids {
		if !want[id] {
			t.Fatalf("unexpected %s", id)
		}
	}
}

func TestExtractCustomEmojiIDsEmpty(t *testing.T) {
	if ids := extractCustomEmojiIDs([]byte(`{"text":"hi"}`)); len(ids) != 0 {
		t.Fatalf("%v", ids)
	}
}
