package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"index/suffixarray"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

func main() {
	searcher := Searcher{}
	err := searcher.Load("completeworks.txt")
	if err != nil {
		log.Fatal(err)
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/search", handleSearch(searcher))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	fmt.Printf("Listening on port %s... \n", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

type SearcherResult struct {
	StartIndex int
	EndIndex   int
	LineNumber int
}
type Searcher struct {
	CompleteWorks      string
	SuffixArray        *suffixarray.Index
	LinesCompleteWorks []string
	MapCompleteWorks   map[int]string
	Result             []SearcherResult
	WordMatches        map[string]bool
}

type IndexLevDist struct {
	Index    int
	Distance int
}

func handleSearch(searcher Searcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := r.URL.Query()["q"]
		if !ok || len(query[0]) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing search query in URL params"))
			return
		}
		results := searcher.Search(query[0])
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		err := enc.Encode(results)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encoding failure"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(buf.Bytes())
	}
}

func (s *Searcher) Load(filename string) error {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Load: %w", err)
	}
	s.CompleteWorks = string(dat)
	s.LinesCompleteWorks = strings.Split(s.CompleteWorks, "\r\n")
	s.SuffixArray = suffixarray.New(dat)
	s.MapCompleteWorks = make(map[int]string)

	for i, line := range s.LinesCompleteWorks {
		s.MapCompleteWorks[i] = line
	}

	return nil
}

func (s *Searcher) Search(query string) []string {
	// idxs := s.SuffixArray.Lookup([]byte(query), -1)
	processedQuery := StringToList(query)

	var wg sync.WaitGroup
	idxs := make([][]IndexLevDist, len(processedQuery))

	for i, word := range processedQuery {
		wg.Add(1)
		go s.WordLookup(word, &idxs[i], &wg)
	}

	wg.Wait()

	sortedIndices := SortIndexList(idxs)
	s.ProcessResults(sortedIndices)
	textResults := s.GatherResultText()
	result := s.ProcessHTMLResult(textResults)

	fmt.Println("Found " + strconv.Itoa(len(result)) + " results.")

	return result
}

func StringToList(query string) []string {
	wordsInQuery := strings.Split(query, " ")
	return wordsInQuery
}

func (s *Searcher) WordLookup(word string, idx *[]IndexLevDist, wg *sync.WaitGroup) {
	defer wg.Done()

	s.WordMatches = make(map[string]bool)

	for lineNumber, line := range s.LinesCompleteWorks {
		processedLine := StringToList(line)
		matchingWords := fuzzy.RankFindNormalizedFold(word, processedLine)
		sort.Sort(matchingWords)

		if len(matchingWords) > 0 {
			currentLine := IndexLevDist{
				lineNumber,
				matchingWords[0].Distance,
			}
			*idx = append(*idx, currentLine)

			for _, matchedWord := range matchingWords {
				s.WordMatches[matchedWord.Target] = true
			}
		}
	}
}

func SortIndexList(idxs [][]IndexLevDist) []int {
	set := make(map[int]bool)

	// Flatten List
	flatList := []IndexLevDist{}

	for _, indexList := range idxs {
		flatList = append(flatList, indexList...)
	}

	// Sort based on Levenshtein Distance
	sort.SliceStable(flatList, func(i, j int) bool {
		return flatList[i].Distance < flatList[j].Distance
	})

	indices := []int{}

	// Convert to "set"
	for _, indexList := range idxs {
		for _, resObj := range indexList {
			if !set[resObj.Index] {
				set[resObj.Index] = true
				indices = append(indices, resObj.Index)
			}
		}
	}

	return indices
}

func (s *Searcher) GetLineStart(resultIndex int) int {

	lineStart := 0
	for curr := resultIndex; curr >= 0; curr-- {
		if s.MapCompleteWorks[curr] == "" {
			lineStart = curr + 1
			break
		}
	}

	return lineStart
}

func (s *Searcher) GetLineEnd(resultIndex int) int {

	lineEnd := 169442
	for curr := resultIndex; curr < len(s.LinesCompleteWorks); curr++ {
		if s.MapCompleteWorks[curr] == "" {
			lineEnd = curr - 1
			break
		}
	}

	return lineEnd
}

func (s *Searcher) ProcessResults(results []int) {

	var processedResults []SearcherResult
	for _, resIndex := range results {
		res := SearcherResult{}
		res.StartIndex = s.GetLineStart(resIndex)
		res.EndIndex = s.GetLineEnd(resIndex)
		res.LineNumber = res.StartIndex

		processedResults = append(processedResults, res)
	}

	s.Result = processedResults
}

func (s *Searcher) GatherResultText() []string {

	var textResult []string
	for _, result := range s.Result {
		currentText := strings.Join(s.LinesCompleteWorks[result.StartIndex:(result.EndIndex+1)], "\r\n")
		textResult = append(textResult, currentText)
	}

	return textResult
}

func (s *Searcher) ProcessHTMLResult(textResults []string) []string {
	resultHTMLReady := make([]string, len(textResults))

	for i, stringLine := range textResults {
		lineHTML := "<i>Line Number: " + strconv.Itoa(s.Result[i].LineNumber) + "</i><br>" + strings.Replace(stringLine, "\r\n", "<br>", -1)

		for matchedWord := range s.WordMatches {
			start := strings.Index(lineHTML, matchedWord)

			if start == -1 {
				continue
			}

			end := start + len(matchedWord)
			lineHTML = lineHTML[:start] + "<mark>" + lineHTML[start:end] + "</mark>" + lineHTML[end:]
		}

		resultHTMLReady[i] = lineHTML

	}

	return resultHTMLReady
}
