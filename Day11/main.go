package main

import (
	"Day11/bst"
	"Day11/trie"
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type Correction struct {
	Word        string   `json:"original"`
	Suggestions []string `json:"suggestions"`
}

type SpellCheckReport struct {
	TotalWords      int          `json:"total_words"`
	MisspelledCount int          `json:"misspelled_count"`
	Corrections     []Correction `json:"corrections"`
}

func main() {

	tr := trie.NewTrie()

	// Load dictionary
	dictFile, err := os.Open("words.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer dictFile.Close()

	scanner := bufio.NewScanner(dictFile)
	for scanner.Scan() {

		tr.Insert(strings.TrimSpace(scanner.Text()))
	}

	start := time.Now()

	// Read input file
	inputFile := "input.txt"
	inputData, err := os.ReadFile(inputFile)
	if err != nil {
		log.Fatal(err)
	}

	inputStr := string(inputData)
	// Handle literal \n sequences in input file (like in line 370)
	inputStr = strings.ReplaceAll(inputStr, `\n`, " ")

	words := strings.Fields(inputStr)
	report := SpellCheckReport{
		TotalWords: len(words),
	}

	for _, w := range words {
		orig := w
		w = strings.TrimSpace(w)
		if w == "" {
			continue
		}

		if !tr.Search(w) {
			report.MisspelledCount++

			// Use BST for suggestions
			suggestionsBST := bst.NewBST()
			candidates := tr.DidYouMean(w, 2)
			for _, c := range candidates {
				dist := editDistance(w, c)
				freq := tr.GetFreq(c)
				suggestionsBST.Insert(c, dist, freq)
			}

			suggestions := suggestionsBST.GetSuggestions()
			report.Corrections = append(report.Corrections, Correction{
				Word:        orig,
				Suggestions: suggestions,
			})
		}
	}

	duration := time.Since(start)
	fmt.Println("Spell check completed in:", duration)

	// Output report
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile("report.json", jsonData, 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Spell check complete. Misspelled: %d/%d\n", report.MisspelledCount, report.TotalWords)

	var autoComplet string
	fmt.Println("Enter a prefix to autocomplete:")
	fmt.Scan(&autoComplet)
	fmt.Println(tr.AutoComplete(autoComplet, 5))
}

func editDistance(s1, s2 string) int {
	m := len(s1)
	n := len(s2)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
		dp[i][0] = i
	}
	for j := range dp[0] {
		dp[0][j] = j
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if s1[i-1] == s2[j-1] {
				dp[i][j] = dp[i-1][j-1]
			} else {
				dp[i][j] = 1 + min(dp[i-1][j], dp[i][j-1], dp[i-1][j-1])
			}
		}
	}
	return dp[m][n]
}

func min(a, b, c int) int {
	if a < b && a < c {
		return a
	}
	if b < c {
		return b
	}
	return c
}
