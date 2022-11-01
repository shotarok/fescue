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
	tPlusOne := t.Add(time.Duration(1) * time.Hour * 24)

	resAtT, err := getLatestRead(f, t)
	if err != nil {
		fmt.Printf("Failed to get the latest read articles: %v", err)
		return
	}
	resAtTPlusOne, err := getLatestRead(f, tPlusOne)
	if err != nil {
		fmt.Printf("Failed to get the latest read articles: %v", err)
	}

	p := func(r *feedly.MarkerLatestReadResponse) {
		b, err := json.MarshalIndent(r, "", "    ")
		if err != nil {
			fmt.Printf("Failed to marshal latestReadRes: %v", err)
		}
		fmt.Println(string(b))
	}
	p(resAtT)
	p(resAtTPlusOne)

	settPlusOne := hashset.New[string]()
	for _, v := range resAtTPlusOne.Entries {
		settPlusOne.Add(v)
	}

	diff := hashset.New[string]()
	for _, v := range resAtT.Entries {
		if !settPlusOne.Contains(v) {
			diff.Add(v)
		}
	}

	fmt.Printf("Read Article Count on %s: %d\n", t.Format("2006-01-02"), diff.Size())
	for i, v := range diff.Values() {
		fmt.Printf("entry %d: %s\n", i, v)
	}
}
