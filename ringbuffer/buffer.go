// Craig Tomkow
// July 30, 2019

package ringbuffer

import (
	"time"
)

// ##### structs #####

type RingBuffer struct {
	// size of the ring buffer, defined by the max backups specified in conf.json
	size int

	// ring buffer is an array of structs [databaseName string, timestamp time.Time]
	// It has an artificial buffer limit size of 31, regardless of user specified buffer size
	//   used as a safeguard to ensure the app doesn't run away with backups filling up the system
	ring [31]struct {
		name      string
		timestamp time.Time
	}

	// the beginning and end pointers of the buffer
	head int
	tail int
}

func (rb *RingBuffer) Initialize(size int, dbName string, existingFilesAsTimestamps []time.Time) []time.Time {

	rb.size = size
	rb.head = 0
	rb.tail = 0 // required for removing the last backup before it falls off the buffer

	var markedForDeletion []time.Time

	for _, elem := range existingFilesAsTimestamps {
		deleteItem := rb.Add(dbName, elem)

		if !deleteItem.IsZero() {
			markedForDeletion = append(markedForDeletion, deleteItem)
		}
	}

	return markedForDeletion
}

func (rb *RingBuffer) Add(dbName string, fileTimestamp time.Time) time.Time {

	var timestampToDelete time.Time
	// check if ring buffer element is empty
	if !rb.ring[rb.head].timestamp.IsZero() {
		timestampToDelete = rb.ring[rb.head].timestamp
	}

	// add to ring buffer
	rb.ring[rb.head] = struct {
		name      string
		timestamp time.Time
	}{name: compileFilename(dbName, fileTimestamp), timestamp: fileTimestamp}

	// in this order!
	rb.updateHead()
	rb.updateTail()

	return timestampToDelete
}

func (rb *RingBuffer) updateHead() {

	rb.head = mod(rb.head+1, rb.size)
}

func (rb *RingBuffer) updateTail() {

	if (rb.tail + 1) == rb.size {
		rb.tail = mod(rb.tail+1, rb.size)
	} else if rb.head == rb.tail {
		rb.tail++
	}
}

func mod(a, b int) int {

	m := a % b
	if a < 0 && b < 0 {
		m -= b
	}
	if a < 0 && b > 0 {
		m += b
	}

	return m
}
