package fixtures

import (
	"github.com/agnivade/levenshtein"
	"strings"
)

type GameType struct {
	homeTeam    string
	awayTeam    string
	homeForm    string
	awayForm    string
	scorePredic string
	gameOdds    string
	gid         string
}

func stringRatio(str1, str2 string) float64 {
	distance := levenshtein.ComputeDistance(str1, str2)
	ratio := float64((len(str1) + len(str2) - distance)) / float64((len(str1) + len(str2)))
	return ratio
}

func isAbbrev(abbrev, str string) bool {
	abb := strings.ToLower(abbrev)
	text := strings.ToLower(str)
	words := strings.Fields(text)
	if abb == "" {
		return true
	} else if abb != "" && text == "" {
		return false
	} else if abb[0:1] != text[0:1] {
		return false
	} else {
		plc := false
		for i, _ := range words[0] {
			if isAbbrev(abb[1:], text[i+1:]) {
				plc = true
				break
			}
		}
		val := isAbbrev(abb[1:], strings.Join(words[1:], " "))

		value := val || plc
		return value
	}
}

func matchGame(t1, t2 string) bool {
	if (t1 == t2) || (stringRatio(t1, t2) >= 0.9) || isAbbrev(t1, t2) || isAbbrev(t2, t1) {
		return true
	} else {
		return false
	}
}

func FindFixture(glist [8][]GameType, hteam string, ateam string) GameType {
	var gm GameType
	for _, h := range glist {
		for _, g := range h {
			if matchGame(g.homeTeam, hteam) && matchGame(g.awayTeam, ateam) {
				gm = g
			}
		}
	}
	return gm
}
