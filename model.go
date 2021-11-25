package lyricsapi

type ColorLyrics struct {
	Lyrics          *LyricsInfo `json:"lyrics"`
	Colors          *ColorsInfo `json:"colors"`
	HasVocalRemoval bool        `json:"hasVocalRemoval"`
}

type LyricsInfo struct {
	// SyncType indicates whether the lyrics are timesynced
	// Possible values: "LINE_SYNCED", "UNSYNCED" (more?)
	SyncType string        `json:"syncType"`
	Lines    []*LyricsLine `json:"lines"`

	Provider            string `json:"provider"`
	ProviderLyricsID    string `json:"providerLyricsId"`
	ProviderDisplayName string `json:"providerDisplayName"`
	SyncLyricsURI       string `json:"syncLyricsUri"`

	IsDenseTypeface  bool   `json:"isDenseTypeface"`
	Language         string `json:"language"`
	IsRtlLanguage    bool   `json:"isRtlLanguage"`
	FullscreenAction string `json:"fullscreenAction"`
	// TODO: find out what the hell alternatives are
	// Alternatives        []interface{} `json:"alternatives"`
}

type ColorsInfo struct {
	Background    int `json:"background"`
	Text          int `json:"text"`
	HighlightText int `json:"highlightText"`
}
type LyricsLine struct {
	// StartTimeMs is the start time of the line (in ms).
	// I have no idea why this is string.
	// "0" if lyrics are not timesynced.
	StartTimeMs string `json:"startTimeMs"`
	Words       string `json:"words"`
	// TODO: find out what the hell syllables are
	// Syllables   []interface{} `json:"syllables"`
}
