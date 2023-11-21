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
}
