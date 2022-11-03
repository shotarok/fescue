package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/daichi-m/go18ds/sets/hashset"
	"github.com/sfanous/go-feedly/feedly"
	feedlytime "github.com/sfanous/go-feedly/pkg/time"
	"golang.org/x/oauth2"
)

// The Feedly API deals with entries older than 31 days are automatically marked as read according to
// https://developer.feedly.com/v3/markers/. Therefore, this script will fetch data in the only last 30 days.
const DAYS_TO_FETCH = 30

func readToken(filename string) (*oauth2.Token, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	oauth2Token := oauth2.Token{}

	if err := json.Unmarshal(b, &oauth2Token); err != nil {
		return nil, err
	}

	return &oauth2Token, nil
}

func readReadArticleCount(filename string) (map[string]int, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var m map[string]int
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func getLatestRead(f *feedly.Client, d time.Time) (res *feedly.MarkerLatestReadResponse, err error) {
	opt := &feedly.MarkerLatestReadOptionalParams{NewerThan: &feedlytime.Time{d}}
	latestReadRes, _, err := f.Markers.LatestRead(opt)
	if err != nil {
		return nil, err
	}
	return latestReadRes, nil
}

func setDiff[T comparable](lhs *hashset.Set[T], rhs *hashset.Set[T]) *hashset.Set[T] {
	s := hashset.New[T]()
	for _, v := range lhs.Values() {
		if !rhs.Contains(v) {
			s.Add(v)
		}
	}

	for _, v := range rhs.Values() {
		if !lhs.Contains(v) {
			s.Add(v)
		}
	}
	return s
}

func main() {
	var (
		tokenFilename       string
		date                string
		readArticleFilename string
	)

	flag.StringVar(&tokenFilename, "token", "", "token persistent store path")
	flag.StringVar(&readArticleFilename, "data", "", "read article count json path")
	flag.StringVar(&date, "date", "", "date (YYYY-MM-DD) when to count read articles")
	flag.Parse()

	oauth2Token, err := readToken(tokenFilename)
	if err != nil {
		fmt.Printf("Failed to fetch OAuth2 token: %v", err)
		return
	}

	m, err := readReadArticleCount(readArticleFilename)
	if err != nil {
		fmt.Printf("Failed to read the read article count json: %v", err)
		return
	}

	f := feedly.NewClient(oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(oauth2Token)))
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		fmt.Printf("Failed to parse a given arg: %v", err)
		return
	}

	var prev *hashset.Set[string]
	for i := 0; i < DAYS_TO_FETCH; i++ {
		d := t.Add(time.Duration(-i) * time.Hour * 24)
		res, err := getLatestRead(f, d)
		if err != nil {
			fmt.Printf("Failed to get the latest read articles: %v", err)
			return
		}
		cur := hashset.New(res.Entries...)
		if prev != nil {
			diff := setDiff(cur, prev)
			m[d.Format("2006-01-02")] = diff.Size()
		}
		prev = cur
	}

	b, err := json.MarshalIndent(m, "", "    ")
	if err != nil {
		fmt.Printf("Failed to marshal latestReadRes: %v", err)
		return
	}
	if err = os.WriteFile(readArticleFilename, b, 0660); err != nil {
		fmt.Printf("Failed to update json: %v", err)
		return
	}
	fmt.Println(string(b))
}
