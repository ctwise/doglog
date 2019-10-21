package main

// Print out the log messages that match the search criteria.
func commandListMessages(opts *options, startingId string) ([]logMessage, string) {
	messages, nextId := fetchMessages(opts, startingId)

	return messages, nextId
}

func printMessages(messages []logMessage, opts *options, nextId string) {
	for _, msg := range messages {
		printMessage(opts, msg)
	}
}
