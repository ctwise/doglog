package main

// Print out the log messages that match the search criteria.
func commandListMessages(opts *options) bool {
	found := false
	for {
		messages, nextId := fetchMessages(opts, "")
		if len(messages) > 0 {
			found = true
			printMessages(messages, opts)
		}
		if len(nextId) == 0 {
			break
		} else {
			delayForSeconds(0.2)
		}
	}

	return found
}

func printMessages(messages []logMessage, opts *options) {
	for _, msg := range messages {
		printMessage(opts, msg)
	}
}
