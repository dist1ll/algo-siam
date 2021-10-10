package csgo

import (
	"github.com/PuerkitoBio/goquery"
	"time"
)

type Team struct {
	Name string
	// HLTV ID for team
	ID int
}

type Event struct {
	// E.g. "IEM Fall 2021 Europe"
	Name string
	// HLTV ID for event
	ID      int
	LogoURL string
}

type Match struct {
	ID     int
	Team1  Team
	Team2  Team
	Date   time.Time
	Event  Event
	Format string
	Result Result
}

type Result struct {
	// Winning team's name e.g. "OG", "Astralis"
	Winner string
	// Numbered score (e.g. "1-0", "3-2"). Winner's score is always listed first.
	Score string
}

type HLTV struct {
	UpcomingPage *goquery.Document
	ResultsPage  *goquery.Document
}
