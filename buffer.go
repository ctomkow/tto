// Craig Tomkow
// July 30, 2019

package main

type CircularQueue struct {
	// size of the queue, defined by the max_backups specified in conf.json
	size int

	// circular queue is an array of structs queue{dbName string, timestamp time.Time}
	// it has an artificial buffer limit size of 31, regardless of user specified max_backups
	queue [31]struct {
		name      string
	}

	// the start and end pointers of the queue
	head int
	tail int
}

func (cq *CircularQueue) Make(size int) {

	cq.size = size
	cq.head = 0
	cq.tail = 0
}

func (cq *CircularQueue) Populate(dataToPopulate []string) []string {

	var bufOverwriteNames []string

	for _, elem := range dataToPopulate {
		bufOverwriteName := cq.Enqueue(elem)

		if bufOverwriteName != "" {
			bufOverwriteNames = append(bufOverwriteNames, bufOverwriteName)
		}
	}

	return bufOverwriteNames
}

func (cq *CircularQueue) Enqueue(element string) string {

	var bufOverwriteName string
	// check if buffer element is not empty
	if cq.queue[cq.head].name != "" {
		bufOverwriteName = cq.queue[cq.head].name
	}

	// add to queue
	cq.queue[cq.head] = struct {
		name      string
	}{name: element}

	// in this order!
	cq.updateHead()
	cq.updateTail()

	return bufOverwriteName
}

func (cq *CircularQueue) updateHead() {

	cq.head = mod(cq.head+1, cq.size)
}

func (cq *CircularQueue) updateTail() {

	if (cq.tail + 1) == cq.size {
		cq.tail = mod(cq.tail+1, cq.size)
	} else if cq.head == cq.tail {
		cq.tail++
	}
}

// simple integer modulo helper function
func mod(firstInt, secondInt int) int {

	modulo := firstInt % secondInt
	if firstInt < 0 && secondInt < 0 {
		modulo -= secondInt
	}
	if firstInt < 0 && secondInt > 0 {
		modulo += secondInt
	}

	return modulo
}
