package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
	"time"
)

const greyEsc = "\033[37m"
const redEsc = "\033[91m"
const greenEsc = "\033[92m"
const yellowEsc = "\033[93m"
const blueEsc = "\033[94m"
const magentaEsc = "\033[95m"
const cyanEsc = "\033[96m"
const whiteEsc = "\033[97m"

//const resetEsc = "\033[39;49m"
const resetEsc = "\033[39;49m"

const debugEsc = blueEsc
const errorEsc = redEsc
const infoEsc = greenEsc
const warnEsc = yellowEsc

const debugLevel = "DEBUG"
const errorLevel = "ERROR"
const fatalLevel = "FATAL"
const infoLevel = "INFO"
const traceLevel = "TRACE"
const warnLevel = "WARN"

const longTimeFormat = "2006-01-02T15:04:05.000Z"

// Format a log message into JSON.
func formatJson(msg logMessage) string {
	var text string

	buf, _ := json.Marshal(msg.fields)
	text = strings.TrimRight(string(buf), "}")
	buf, _ = json.Marshal(msg.tags)
	text += ",\"tags\":"
	text += string(buf)
	text += "}"

	text = strings.ReplaceAll(text, "\\\"", "\"")

	return text
}

// Print a single log message
func printMessage(opts *options, msg logMessage) {
	adjustMessage(msg, opts.color)

	var text string

	if opts.json {
		text = msg.fields[jsonField]
	} else {
		for _, f := range opts.serverConfig.Formats() {
			text = tryFormat(msg, f.Name, f.Format)
			if len(text) > 0 {
				break
			}
		}
	}

	if len(text) > 0 {
		if strings.HasPrefix(text, "No Formats Defined>>") {
			fmt.Println("stop")
		}
		fmt.Println(text)
	} else {
		// Last case fallback in case none of the formats (including the default) match
		text = msg.fields[jsonField]
		fmt.Println(text)
	}
}

// Try to apply a format template.
// returns: empty string if the format failed.
func tryFormat(msg logMessage, tmplName string, tmpl string) string {
	funcMap := template.FuncMap{
		"ToUpper": strings.ToUpper,
		"ToLower": strings.ToLower,
	}
	var t = template.Must(template.New(tmplName).Option("missingkey=error").Funcs(funcMap).Parse(tmpl))
	var result bytes.Buffer

	if err := t.Execute(&result, msg.fields); err == nil {
		return result.String()
	}

	return ""
}

// Convert a timestamp to a long time string.
func longTime(t time.Time) string {
	t = t.In(time.Local)
	return t.Format(longTimeFormat)
}

// "Cleanup" the log message and add helper fields.
func adjustMessage(msg logMessage, isTty bool) {
	requestPage := msg.fields[requestPageField]
	if len(requestPage) > 1 && !strings.HasPrefix(requestPage, "/") {
		msg.fields[requestPageField] = "/" + requestPage
	}

	originalMessage := msg.fields[originalMessageField]
	if len(originalMessage) == 0 {
		originalMessage = msg.fields[fullMessageField]
		msg.fields[originalMessageField] = originalMessage
	}

	timestamp := msg.timestamp
	msg.fields[longTimestampField] = longTime(timestamp)

	classname := msg.fields[classnameField]
	if len(classname) > 0 {
		msg.fields[shortClassnameField] = createShortClassname(classname)
	}

	level := normalizeLevel(msg)

	constructMessageText(msg, originalMessage)

	setupColors(isTty, level, msg)
}

// Setup the colors in the message structure.
func setupColors(isTty bool, level string, msg logMessage) {
	if isTty {
		computeLevelColor(level, msg)
		// Add color escapes
		msg.fields[blueField] = blueEsc
		msg.fields[redField] = redEsc
		msg.fields[greenField] = greenEsc
		msg.fields[yellowField] = yellowEsc
		msg.fields[greyField] = greyEsc
		msg.fields[whiteField] = whiteEsc
		msg.fields[cyanField] = cyanEsc
		msg.fields[magentaField] = magentaEsc
		msg.fields[resetField] = resetEsc
	} else {
		// Add color escapes
		msg.fields[blueField] = ""
		msg.fields[redField] = ""
		msg.fields[greenField] = ""
		msg.fields[yellowField] = ""
		msg.fields[greyField] = ""
		msg.fields[whiteField] = ""
		msg.fields[cyanField] = ""
		msg.fields[magentaField] = ""
		msg.fields[levelColorField] = ""
		msg.fields[resetField] = ""
	}
}

// Construct the "best" version of the log messages main text. This will look in multiple fields, attempt to
// append multi-line text (stacktraces) onto the message text, etc.
func constructMessageText(msg logMessage, originalMessage string) {
	const nestedException = "; nested exception "
	const newlineNnestedException = ";\nnested exception "

	messageText := msg.fields[messageField]
	if len(messageText) == 0 {
		messageText = originalMessage
	}
	if strings.Contains(messageText, nestedException) {
		messageText = strings.Replace(messageText, nestedException, newlineNnestedException, -1)
	}
	if len(originalMessage) > 0 && messageText != originalMessage {
		extraInfo := strings.Split(originalMessage, "\n")
		if len(extraInfo) == 2 {
			messageText = messageText + "\n" + extraInfo[1]
		}
		if len(extraInfo) > 2 {
			messageText = messageText + "\n" + strings.Join(extraInfo[1:len(extraInfo)-1], "\n")
		}
	}
	msg.fields[jsonField] = formatJson(msg)
	if len(messageText) == 0 {
		messageText = msg.fields[jsonField]
	}
	// Replace \" with plain "
	messageText = strings.ReplaceAll(messageText, "\\\"", "\"")
	msg.fields[messageTextField] = messageText
}

// Normalize the "level" of the message.
func normalizeLevel(msg logMessage) string {
	level := msg.fields[levelField]
	if len(level) == 0 {
		level = msg.fields[logLevelField]
		delete(msg.fields, logLevelField)
	}
	if len(level) == 0 {
		level = msg.fields[statusField]
		delete(msg.fields, statusField)
	}
	if len(level) == 0 {
		level = msg.fields[logStatusField]
		delete(msg.fields, logStatusField)
	}
	level = strings.ToUpper(level)
	if strings.HasPrefix(level, "E") {
		level = errorLevel
	} else if strings.HasPrefix(level, "F") {
		level = fatalLevel
	} else if strings.HasPrefix(level, "I") {
		level = infoLevel
	} else if strings.HasPrefix(level, "W") {
		level = warnLevel
	} else if strings.HasPrefix(level, "D") {
		level = debugLevel
	} else if strings.HasPrefix(level, "T") {
		level = traceLevel
	}
	msg.fields[levelField] = level
	return level
}

// Compute the color that should be used to display the log level in the message output.
func computeLevelColor(level string, msg logMessage) {
	var levelColor string
	switch level {
	case debugLevel, traceLevel:
		levelColor = debugEsc
	case infoLevel:
		levelColor = infoEsc
	case warnLevel:
		levelColor = warnEsc
	case errorLevel, fatalLevel:
		levelColor = errorEsc
	}
	if len(levelColor) > 0 {
		msg.fields[levelColorField] = levelColor
	} else {
		msg.fields[levelColorField] = ""
	}
}

// Create a shortened version of the Java classname.
func createShortClassname(classname string) string {
	parts := strings.Split(classname, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return classname
}
