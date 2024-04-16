package lyrics

import "encoding/json"

type LyricsResult struct {
	Lyrics LyricsInfo `json:"lyrics"`
	Colors ColorsInfo `json:"colors"`
}

type LyricsInfo struct {
	SyncType string       `json:"syncType"`
	Lines    []LyricsLine `json:"lines"`
}

type LyricsLine struct {
	Time  int    `json:"startTimeMs,string"`
	Words string `json:"words"`
}

func (l *LyricsLine) MarshalJSON() ([]byte, error) {
	type alias struct {
		Time  int    `json:"time"`
		Words string `json:"words"`
	}
	return json.Marshal(alias(*l))
}

type ColorsInfo struct {
	Background    int `json:"background"`
	Text          int `json:"text"`
	HighlightText int `json:"highlightText"`
	// add 0x1000000 to this value to get the color. currently
	// only background is dynamic; text and highlightText
	// contain constant values -16777216 and -1 (black and white)
}

type StateResult struct {
	Progress int       `json:"progress_ms"`
	Playing  bool      `json:"is_playing"`
	Item     *struct { // nil == nothing playing
		ID       string `json:"id"`
		Name     string `json:"name"`
		Duration int    `json:"duration_ms"`
	} `json:"item"`
}
