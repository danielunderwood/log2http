package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/bits-and-blooms/bloom/v3"
)

type Deduplicator interface {
	Exists([]byte) (bool, error)
	Add([]byte) error
}

// The null deduplicator doesn't do any deduplication and always returns
// false for existence
type NullDeduplicator struct{}

func NewNullDeduplicator() *NullDeduplicator {
	return &NullDeduplicator{}
}

func (*NullDeduplicator) Exists([]byte) (bool, error) {
	return false, nil
}

func (*NullDeduplicator) Add([]byte) error {
	return nil
}

// The bloom filter deduplicator does deduplication based on a bloom filter
type BloomFilterDeduplicator struct {
	path string
	bf   *bloom.BloomFilter
}

// Capacity for size n with a given false positive chance
func NewBloomFilterDeduplicator(path string, n uint, errorRate float64) *BloomFilterDeduplicator {
	if _, err := os.Stat(path); err == nil {
		// File exists
		// Note that in this case, the arguments are ignored
		file, err := os.Open(path)
		if err != nil {
			fmt.Printf("ERROR: Could not open bloom file: %s\n", err)
			return nil
		}
		defer file.Close()
		// Note that the parameters won't be preserved
		bf := bloom.NewWithEstimates(n, errorRate)
		_, err = bf.ReadFrom(file)
		if err == io.EOF {
			// If we get an EOF, it means the old file was corrupted, so we can just create a new one
			fmt.Printf("WARNING: Bloom filter file is corrupted. Creating a new one\n")
			bf := bloom.NewWithEstimates(n, errorRate)
			return &BloomFilterDeduplicator{path: path, bf: bf}
		} else if err != nil {
			fmt.Printf("ERROR: Could not read bloom file: %s\n", err)
			return nil
		}
		return &BloomFilterDeduplicator{path: path, bf: bf}
	} else if errors.Is(err, os.ErrNotExist) {
		// Doesn't exist, need to create new filter
		bf := bloom.NewWithEstimates(n, errorRate)
		return &BloomFilterDeduplicator{path: path, bf: bf}
	} else {
		// Didn't do what we expected
		fmt.Printf("ERROR: Could not process bloom path: %s\n", err)
		return nil
	}
}

func (d *BloomFilterDeduplicator) Exists(b []byte) (bool, error) {
	return d.bf.Test(b), nil
}

func (d *BloomFilterDeduplicator) Add(b []byte) error {
	d.bf = d.bf.Add(b)
	// Save the filter
	// This probably isn't the best way to do this and could lead to a number of problems,
	// but it will work for now
	// It would probably be better to do this on a separate thread at some fixed interval

	// Use a temporary file and then move it to the chosen path to do an atomic write
	// This prevents corruption of the file in the case that the program dies while writing the file
	file, err := os.CreateTemp("", "*")
	if err != nil {
		fmt.Printf("WARNING: Could not open bloom file for saving: %s\n", err)
	}
	defer file.Close()

	_, err = d.bf.WriteTo(file)
	if err != nil {
		fmt.Printf("WARNING: Unable to write bloom file for saving: %s\n", err)
	} else {
		os.Rename(file.Name(), d.path)
	}

	return nil
}
