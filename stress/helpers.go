package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
)

var httpUrl = os.Getenv("HTTP_URL")

func Post(client *http.Client, route string, query any, result any) error {
	body, err := json.Marshal(query)
	if err != nil {
		return err
	}
	response, err := client.Post(
		httpUrl+route,
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}
	defer response.Body.Close()
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return err
	}
	return nil
}

func RandomInt(min int, max int) int {
	return min + rand.Intn(max+1-min)
}

func ListRandomSubsets(items []string) [][]string {
	subsets := ListSubsets(items)
	rand.Shuffle(len(subsets), func(i, j int) {
		subsets[i], subsets[j] = subsets[j], subsets[i]
	})
	return subsets
}

func ListSubsets(items []string) [][]string {
	n := len(items)
	total := 1 << n
	subsets := make([][]string, 0, total)

	for mask := 0; mask < total; mask++ {
		subset := make([]string, 0)
		for i := 0; i < n; i++ {
			if mask&(1<<i) != 0 {
				subset = append(subset, items[i])
			}
		}
		subsets = append(subsets, subset)
	}
	return subsets
}

func ListDifference(a, b []string) []string {
	m := make(map[string]bool)
	for _, x := range b {
		m[x] = true
	}
	var diff []string
	for _, x := range a {
		if _, ok := m[x]; !ok {
			diff = append(diff, x)
		}
	}
	return diff
}
