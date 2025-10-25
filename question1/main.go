package main

import (
	"fmt"
	"sync"
	"time"
)

// calculateEvenSum processes a slice chunk and sends the sum of even numbers to the results channel
func calculateEvenSum(numbers []int, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	
	sum := 0
	for _, num := range numbers {
		if num%2 == 0 {
			sum += num
		}
	}
	
	results <- sum
}

// sumEvenNumbersConcurrent divides the slice among workers and calculates sum concurrently
func sumEvenNumbersConcurrent(numbers []int, numWorkers int) int {
	if len(numbers) == 0 {
		return 0
	}
	
	// Create channel with capacity equal to number of workers
	results := make(chan int, numWorkers)
	var wg sync.WaitGroup
	
	// use chunk so each worker not process entire numbers
	chunkSize := len(numbers) / numWorkers
	//The remainder tells you how many workers should get one extra item so all items are processed.
	remainder := len(numbers) % numWorkers

	startIdx := 0
	for i := 0; i < numWorkers; i++ {
		// Adjust chunk size to distribute remainder
		currentChunkSize := chunkSize
		if i < remainder {
			currentChunkSize++
		}

		// Calculate end index
		endIdx := startIdx + currentChunkSize
		if endIdx > len(numbers) {
			endIdx = len(numbers)
		}

		// Skip if no elements to process
		if startIdx >= len(numbers) {
			break
		}

		//log.Println(numbers[startIdx:endIdx],remainder,currentChunkSize)

		// Launch goroutine for this chunk
		wg.Add(1)
		go calculateEvenSum(numbers[startIdx:endIdx], results, &wg)

		startIdx = endIdx
	}

	// Close results channel when all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()
	
	// Collect results from all workers
	totalSum := 0
	for partialSum := range results {
		totalSum += partialSum
	}
	
	return totalSum
}

// main is the entry point of the application, demonstrating concurrent and sequential even-number summation.
func main() {
	// Create a large slice of integers for testing
	const sliceSize = 1000000
	numbers := make([]int, sliceSize)
	for i := 0; i < sliceSize; i++ {
		numbers[i] = i + 1
	}

	// Test with 4 workers
	numWorkers := 4
	
	fmt.Printf("Processing %d numbers with %d workers...\n", sliceSize, numWorkers)
	
	startTime := time.Now()
	sum := sumEvenNumbersConcurrent(numbers, numWorkers)
	duration := time.Since(startTime)

	fmt.Printf("Sum of all even numbers: %d\n", sum)
	fmt.Printf("Time taken: %v\n", duration)

	// Verify with sequential calculation
	fmt.Println("\nVerifying with sequential calculation...")
	startTime = time.Now()
	expectedSum := 0
	for _, num := range numbers {
		if num%2 == 0 {
			expectedSum += num
		}
	}
	normalDuration := time.Since(startTime)

	fmt.Printf("Expected sum: %d\n", expectedSum)
	fmt.Printf("Sequential time: %v\n", normalDuration)

	if sum == expectedSum {
		fmt.Println("\n✓ Result verified successfully!")
	} else {
		fmt.Println("\n✗ Result mismatch!")
	}
}
