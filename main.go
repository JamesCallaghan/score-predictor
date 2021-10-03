package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/JamesCallaghan/score-predictor/pkg/fixtures"
	"github.com/PuerkitoBio/goquery"
	"gopkg.in/urfave/cli.v1"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func FileExists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func main() {

	var request [8]*http.Request
	var err [8]error
	var response [8]*http.Response
	var document [8]*goquery.Document
	var url [8]string
	var gameWeek [8][]GameType
	var fixture [8]GameType
	var hTeam [8]string
	var aTeam [8]string
	var scorePrediction [8]string
	var homeForm [8]string
	var awayForm [8]string
	var matchOdds [8]string
	var gameID [8]string
	var err2 error
	var leaguePos string
	var leaguePositions []string
	var existingFileName [8]string
	var localFilesExist [8]bool

	app := cli.NewApp()
	app.Name = "Score Predictor CLI"
	app.Usage = "Print score predictions"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "home, t",
			Value: "",
			Usage: "Home team",
		},
		cli.StringFlag{
			Name:  "away, a",
			Value: "",
			Usage: "Away team",
		},
		cli.BoolFlag{
			Name:  "delete, d",
			Usage: "Delete local html files",
		},
	}
	app.Action = func(c *cli.Context) error {

		currentDate := time.Now().Local()

		if c.Bool("delete") {
			for a := 0; a < 8; a++ {
				existingFileName[a] = currentDate.AddDate(0, 0, a).Format("20060102") + ".html"
				localFilesExist[a], _ = FileExists(existingFileName[a])
				if localFilesExist[a] == true {
					err2 := os.Remove(existingFileName[a])

					if err2 != nil {
						fmt.Println(err2)
					}
				}
			}
		}

		// Create HTTP client with timeout
		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		for a := 0; a < 8; a++ {
			newDate := currentDate.AddDate(0, 0, a)
			url[a] = "https://www.predictz.com/predictions/" + newDate.Format("20060102") + "/"
			localFileExists, _ := FileExists(newDate.Format("20060102") + ".html")
			if localFileExists == false {
				// Create and modify HTTP request before sending
				request[a], err[a] = http.NewRequest("GET", "https://www.predictz.com/predictions/"+currentDate.AddDate(0, 0, a).Format("20060102")+"/", nil)
				if err[a] != nil {
					log.Fatal(err[a])
				}
				request[a].Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.182 Safari/537.36")

				// Make request
				response[a], err[a] = client.Do(request[a])
				if err[a] != nil {
					log.Fatal(err[a])
				}
				defer response[a].Body.Close()

				var buf bytes.Buffer
				tee := io.TeeReader(response[a].Body, &buf)

				// Create output file
				outFile, err := os.Create(newDate.Format("20060102") + ".html")
				if err != nil {
					log.Fatal(err)
				}
				defer outFile.Close()

				// Copy data from HTTP response to file
				_, err2 = io.Copy(outFile, tee)
				if err != nil {
					log.Fatal(err2)
				}
				defer response[a].Body.Close()

				// Create a goquery document from the HTTP response
				document[a], err2 = goquery.NewDocumentFromReader(&buf)
				if err2 != nil {
					log.Fatal("Error loading HTTP response body. ", err2)
				}
			} else {
				f, e := os.Open(newDate.Format("20060102") + ".html")
				if e != nil {
					log.Fatal(e)
				}
				defer f.Close()

				// Create a goquery document from the HTTP response
				document[a], err2 = goquery.NewDocumentFromReader(f)
				if err2 != nil {
					log.Fatal("Error loading HTTP response body. ", err2)
				}
			}

			gameWeek[a] = make([]GameType, 0)

			document[a].Find(".pttr.ptcnt").Each(func(_ int, game *goquery.Selection) {

				game.Find(".pttd.ptmobh").Each(func(_ int, home *goquery.Selection) {
					hTeam[a] = home.Text()
					fixture[a].homeTeam = hTeam[a]
				})

				game.Find(".pttd.ptmoba").Each(func(_ int, away *goquery.Selection) {
					aTeam[a] = away.Text()
					fixture[a].awayTeam = aTeam[a]
				})

				game.Find(".pttd.ptprd").Each(func(_ int, p *goquery.Selection) {
					p.Find(".ptpredboxsml").Each(func(_ int, prediction *goquery.Selection) {
						scorePrediction[a] = prediction.Text()
						fixture[a].scorePredic = scorePrediction[a]
					})
				})

				game.Find(".pttd.ptlast5h").Each(func(_ int, past_results *goquery.Selection) {
					homeForm[a] = ""
					past_results.Find(".ptneonboxsml2").Each(func(_ int, home_form *goquery.Selection) {
						homeForm[a] = homeForm[a] + home_form.Text()
					})
					fixture[a].homeForm = homeForm[a]
				})

				game.Find(".pttd.ptlast5a").Each(func(_ int, past_results *goquery.Selection) {
					awayForm[a] = ""
					past_results.Find(".ptneonboxsml2").Each(func(_ int, away_form *goquery.Selection) {
						awayForm[a] = awayForm[a] + away_form.Text()
					})
					fixture[a].awayForm = awayForm[a]
				})

				matchOdds[a] = "H/D/A---"
				game.Find(".pttd.ptodds").Each(func(_ int, odds *goquery.Selection) {

					odds.Find("a").Each(func(_ int, odds_string *goquery.Selection) {
						matchOdds[a] = matchOdds[a] + odds_string.Text() + "---"
					})
					fixture[a].gameOdds = matchOdds[a]

				})

				game.Find(".pttd.ptgame").Each(func(_ int, game_link *goquery.Selection) {
					game_link.Find("a").Each(func(_ int, link *goquery.Selection) {
						gameID[a], _ = link.Attr("href")
						fixture[a].gid = gameID[a]
					})

				})

				gameWeek[a] = append(gameWeek[a], fixture[a])

			})

		}

		homeFlag := c.GlobalString("home")
		awayFlag := c.GlobalString("away")
		g1 := fixtures.FindFixture(gameWeek, homeFlag, awayFlag)
		fmt.Printf("Score Prediction is: %s\n", g1.scorePredic)
		fmt.Printf("Odds are: %s\n", g1.gameOdds)
		fmt.Printf("%s past 5 games form: %s\n", homeFlag, g1.homeForm)
		fmt.Printf("%s past 5 games form: %s\n", awayFlag, g1.awayForm)
		fmt.Printf("Game link: %s\n", g1.gid)

		url2 := g1.gid

		request2, err2 := http.NewRequest("GET", url2, nil)
		if err2 != nil {
			log.Fatal(err2)
		}
		request2.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.182 Safari/537.36")

		// Make request
		response2, err2 := client.Do(request2)
		if err2 != nil {
			log.Fatal(err2)
		}
		defer response2.Body.Close()

		document2, err2 := goquery.NewDocumentFromReader(response2.Body)
		if err2 != nil {
			log.Fatal("Error loading HTTP response body. ", err2)
		}

		leaguePositions = make([]string, 0)
		document2.Find(".statboxth.statbox2").Each(func(_ int, posdiv *goquery.Selection) {
			posdiv.Find(".pztablesm.w100p").Each(func(_ int, postable *goquery.Selection) {
				postable.Find("tr").Each(func(_ int, postr *goquery.Selection) {
					pclass, _ := postr.Attr("class")
					if pclass == "pzcnt" {
						postr.Find("td").Each(func(ha int, lpos *goquery.Selection) {
							leaguePos = lpos.Text()
							if strings.Contains(leaguePos, "th") || strings.Contains(leaguePos, "st") || strings.Contains(leaguePos, "nd") || strings.Contains(leaguePos, "rd") {
								leaguePositions = append(leaguePositions, leaguePos)
							}
						})

					}
				})
			})
		})

		fmt.Printf("%s league position: %s\n", homeFlag, leaguePositions[0])
		fmt.Printf("%s league position: %s\n", awayFlag, leaguePositions[1])

		return nil
	}
	app.Run(os.Args)

}
