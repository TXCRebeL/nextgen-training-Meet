package trie

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

func loadTrie(b *testing.B) *Trie {
	tr := NewTrie()

	file, err := os.Open("../words.txt")
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		tr.Insert(strings.TrimSpace(scanner.Text()))
	}

	return tr
}

func BenchmarkDidYouMean(b *testing.B) {
	tr := loadTrie(b)

	misspellings := []string{"mispeleled", "progr", "intrface", "concurnt"}

	for _, m := range misspellings {
		m := m // fix closure issue

		b.Run("DidYouMean_"+m, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = tr.DidYouMean(m, 2)
			}
		})
	}
}

func BenchmarkAutocomplete(b *testing.B) {
	tr := loadTrie(b)

	prefixes := []string{"th", "pro", "concur", "sch"}

	for _, p := range prefixes {
		p := p // fix closure issue

		b.Run("AutoComplete_"+p, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = tr.AutoComplete(p, 5)
			}
		})
	}
}
