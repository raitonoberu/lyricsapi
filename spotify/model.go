package spotify

type LyricsLine struct {
	Time  int    `json:"startTimeMs,string"`
	Words string `json:"words"`
}

type Track struct {
	ID       string   `json:"id"`
	Album    Album    `json:"album"`
	Artists  []Artist `json:"artists"`
	Duration int      `json:"duration_ms"`
	Name     string   `json:"name"`
}

type Artist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Album struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type StateResult struct {
	Progress int    `json:"progress_ms"`
	Playing  bool   `json:"is_playing"`
	Item     *Track `json:"item"`
}
