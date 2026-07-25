package hw03frequencyanalysis

import (
	"sort"
	"strings"
)

func Top10(text string) []string {
	if text == "" {
		return []string{}
	}

	words := strings.Fields(text)

	freqMap := make(map[string]int)
	for _, word := range words {
		freqMap[word]++
	}

	type wordFreq struct {
		word string
		freq int
	}

	wordFreqs := make([]wordFreq, 0, len(freqMap))
	for word, freq := range freqMap {
		wordFreqs = append(wordFreqs, wordFreq{word: word, freq: freq})
	}

	sort.Slice(wordFreqs, func(i, j int) bool {
		if wordFreqs[i].freq != wordFreqs[j].freq {
			return wordFreqs[i].freq > wordFreqs[j].freq
		}
		return wordFreqs[i].word < wordFreqs[j].word
	})

	result := make([]string, 0, 10)
	for i := 0; i < len(wordFreqs) && i < 10; i++ {
		result = append(result, wordFreqs[i].word)
	}

	return result
}
