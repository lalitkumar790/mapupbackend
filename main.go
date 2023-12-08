// main.go

package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"sync"
	"time"
)

// Input structure for the JSON payload
type Input struct {
	ToSort [][]int `json:"to_sort"`
}

// Output structure for the JSON response
type Output struct {
	SortedArrays [][]int `json:"sorted_arrays"`
	TimeNS       int64   `json:"time_ns"`
}

// SequentialSort sorts each sub-array sequentially
func SequentialSort(input []int) []int {
	sort.Ints(input)
	return input
}

// ConcurrentSort sorts each sub-array concurrently using goroutines and channels
func ConcurrentSort(input []int, wg *sync.WaitGroup, ch chan []int) {
	defer wg.Done()
	sort.Ints(input)
	ch <- input
}

// ProcessSingleHandler handles the /process-single endpoint
func ProcessSingleHandler(w http.ResponseWriter, r *http.Request) {
	var input Input
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	startTime := time.Now()
	var sortedArrays [][]int

	for _, arr := range input.ToSort {
		sortedArrays = append(sortedArrays, SequentialSort(arr))
	}

	timeTaken := time.Since(startTime)

	output := Output{
		SortedArrays: sortedArrays,
		TimeNS:       timeTaken.Nanoseconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}

// ProcessConcurrentHandler handles the /process-concurrent endpoint
func ProcessConcurrentHandler(w http.ResponseWriter, r *http.Request) {
	var input Input
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	startTime := time.Now()
	var sortedArrays [][]int
	var wg sync.WaitGroup
	ch := make(chan []int)

	for _, arr := range input.ToSort {
		wg.Add(1)
		go ConcurrentSort(arr, &wg, ch)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for sortedArr := range ch {
		sortedArrays = append(sortedArrays, sortedArr)
	}

	timeTaken := time.Since(startTime)

	output := Output{
		SortedArrays: sortedArrays,
		TimeNS:       timeTaken.Nanoseconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}

func main() {
	http.HandleFunc("/process-single", ProcessSingleHandler)
	http.HandleFunc("/process-concurrent", ProcessConcurrentHandler)

	http.ListenAndServe(":8000", nil)
}
