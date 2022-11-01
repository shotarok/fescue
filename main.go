package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/daichi-m/go18ds/sets/hashset"
	"github.com/sfanous/go-feedly/feedly"
	feedlytime "github.com/sfanous/go-feedly/pkg/time"
	"golang.org/x/oauth2"
)

const DAYS = 30

func fetchToken(filename string) (*oauth2.Token, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	oauth2Token := oauth2.Token{}

	if err := json.Unmarshal(b, &oauth2Token); err != nil {
		return nil, err
	}

	return &oauth2Token, nil
}

func getLatestRead(f *feedly.Client, d time.Time) (res *feedly.MarkerLatestReadResponse, err error) {
	opt := &feedly.MarkerLatestReadOptionalParams{NewerThan: &feedlytime.Time{d}}
	latestReadRes, _, err := f.Markers.LatestRead(opt)
	if err != nil {
		return nil, err
	}
	return latestReadRes, nil
}

func Diff(lhs *hashset.Set[string], rhs *hashset.Set[string]) *hashset.Set[string] {
	s := hashset.New[string]()
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
		filename string
		date     string
	)

	flag.StringVar(&filename, "file", "", "token persistent store path")
	flag.StringVar(&date, "date", "", "date (YYYY-MM-DD) when to count read articles")
	flag.Parse()

	oauth2Token, err := fetchToken(filename)
	if err != nil {
		fmt.Printf("Failed to fetch OAuth2 token: %v", err)

		return
	}

	f := feedly.NewClient(oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(oauth2Token)))
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		fmt.Printf("Failed to parse a given arg: %v", err)
	}

	m := make(map[string]int)
	var prev *hashset.Set[string]
	for i := 0; i < DAYS; i++ {
		d := t.Add(time.Duration(-i) * time.Hour * 24)
		res, err := getLatestRead(f, d)
		if err != nil {
			fmt.Printf("Failed to get the latest read articles: %v", err)
			return
		}
		cur := hashset.New[string](res.Entries...)
		if prev != nil {
			diff := Diff(cur, prev)
			m[d.Format("2006-01-02")] = diff.Size()
		}
		prev = cur
	}

	p := func(r interface{}) {
		b, err := json.MarshalIndent(r, "", "    ")
		if err != nil {
			fmt.Printf("Failed to marshal latestReadRes: %v", err)
		}
		fmt.Println(string(b))
	}
	p(m)
}
