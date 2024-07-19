package lyrics

import "encoding/json"

type LyricsLine struct {
	Time  int    `json:"startTimeMs,string"`
	Words string `json:"words"`
}

type lyricsResult struct {
	Lyrics struct {
		Lines []LyricsLine `json:"lines"`
	} `json:"lyrics"`
}

func (l *LyricsLine) MarshalJSON() ([]byte, error) {
	type alias struct {
		Time  int    `json:"time"`
		Words string `json:"words"`
	}
	return json.Marshal(alias(*l))
}

type Track struct {
	ID         string    `json:"id"`
	Album      Album     `json:"album"`
	Artists    []Artists `json:"artists"`
	DurationMs int       `json:"duration_ms"`
	Name       string    `json:"name"`
}

type Artists struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Album struct {
	ID   string `json:"id"`
	Name string `json:"name"`
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
