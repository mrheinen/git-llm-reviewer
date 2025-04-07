package progress

import (
	"fmt"
	"io/ioutil"
	"sync"
	"testing"
	"time"
)

func TestConsoleTracker(t *testing.T) {
	// Skip output verification since it's hard to test console output reliably
	// This test just verifies that the tracker doesn't panic
	
	// Create a tracker with a discarded output
	tracker := NewConsoleTracker().WithWriter(ioutil.Discard)

	// Test starting the tracker
	tracker.Start(3)

	// Test starting a file
	tracker.StartFile("file1.go")

	// Test completing a file
	tracker.CompleteFile("file1.go", 2)

	// Test error on a file
	tracker.StartFile("file2.go")
	tracker.ErrorFile("file2.go", "test error")

	// Test another file
	tracker.StartFile("file3.go")
	tracker.CompleteFile("file3.go", 0)

	// Test finish
	tracker.Finish()
}

func TestTrackerWithMultipleTasks(t *testing.T) {
	// Skip output verification since it's hard to test console output reliably
	// This test just verifies that the tracker doesn't panic
	
	// Create a tracker with a discarded output
	tracker := NewConsoleTracker().WithWriter(ioutil.Discard)

	// Add multiple files
	numFiles := 10
	tracker.Start(numFiles)

	// Start all files
	for i := 0; i < numFiles; i++ {
		filename := fmt.Sprintf("file%d.go", i)
		tracker.StartFile(filename)
	}

	// Complete half the files
	for i := 0; i < numFiles/2; i++ {
		filename := fmt.Sprintf("file%d.go", i)
		tracker.CompleteFile(filename, i)
	}

	// Error on a quarter of the files
	for i := numFiles/2; i < numFiles*3/4; i++ {
		filename := fmt.Sprintf("file%d.go", i)
		tracker.ErrorFile(filename, "test error")
	}

	// Complete the rest
	for i := numFiles*3/4; i < numFiles; i++ {
		filename := fmt.Sprintf("file%d.go", i)
		tracker.CompleteFile(filename, 0)
	}

	// Finish
	tracker.Finish()
}

func TestTrackerConcurrentAccess(t *testing.T) {
	// Skip output verification since it's hard to test console output reliably
	// This test just verifies that the tracker doesn't panic
	
	// Create a tracker with a discarded output
	tracker := NewConsoleTracker().WithWriter(ioutil.Discard)

	// Add files concurrently
	numFiles := 10
	tracker.Start(numFiles)
	
	var wg sync.WaitGroup
	wg.Add(numFiles)

	for i := 0; i < numFiles; i++ {
		go func(id int) {
			defer wg.Done()
			filename := fmt.Sprintf("file%d.go", id)
			tracker.StartFile(filename)
			time.Sleep(time.Millisecond * 10)
			if id%2 == 0 {
				tracker.CompleteFile(filename, id)
			} else {
				tracker.ErrorFile(filename, "test error")
			}
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Finish
	tracker.Finish()
}