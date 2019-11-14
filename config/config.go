// Package config is a wrapper for an INI configuration file.
// The package is domain-specific, not general purpose.
package config

import (
	"fmt"
	"gopkg.in/ini.v1"
	"os"
	"path/filepath"
	"strings"
)

const NoFormatDefined = "No Formats Defined>>"

const FullMessageField = "full_message"
const LevelField = "level"
const MessageField = "message"
const ClassnameField = "classname"

const formatsSection string = "formats" // [formats]
const serverSection string = "server"   // [server]
const fieldSection string = "fields"    // [fields]

var storedFormats []FormatDefinition = nil // Stores formats so we don't keep re-reading them
var storedFields map[string][]string = nil // Stores field mappings so we don't keep re-reading them

// IniFile is a wrapper around the INI file reader
type IniFile struct {
	ini *ini.File
}

// FormatDefinition stores a single format line.
type FormatDefinition struct {
	Name   string
	Format string
}

// New creates a new INI file reader and wraps it.
func New(configPath string) (*IniFile, error) {
	f, err := readConfig(configPath)
	if err == nil {
		return &IniFile{ini: f}, nil
	} else {
		return nil, err
	}
}

// ApiKey gets the API key from the config file. Defaults to an empty string.
func (c *IniFile) ApiKey() string {
	server := c.ini.Section(serverSection)
	return server.Key("api-key").MustString("")
}

// ApplicationKey gets the application key from the config file. Defaults to an empty string.
func (c *IniFile) ApplicationKey() string {
	server := c.ini.Section(serverSection)
	return server.Key("application-key").MustString("")
}

// Formats gets the log messages formats from the config file. Adds a final default format case so the user knows that
// no formats were applied successfully.
func (c *IniFile) Formats() (formats []FormatDefinition) {
	if storedFormats == nil {
		for _, f := range c.ini.Section(formatsSection).Keys() {
			formats = append(formats, FormatDefinition{Name: f.Name(), Format: f.Value()})
		}
		formats = append(formats, FormatDefinition{Name: "_default", Format: NoFormatDefined + " {{._json}}"})
		storedFormats = formats
	}

	return storedFormats
}

// Fields gets the field mappings from the config file. These will be merged with the defaults.
func (c *IniFile) Fields() (fields map[string][]string) {
	if storedFields == nil {
		storedFields = make(map[string][]string)
		storedFields[LevelField] = []string{"level", "status", "loglevel", "log_status"}
		storedFields[MessageField] = []string{"message", "msg"}
		storedFields[FullMessageField] = []string{"full_message", "original_message"}
		storedFields[ClassnameField] = []string{"logger_name"}
		for _, f := range c.ini.Section(fieldSection).Keys() {
			name := f.Name()
			value := f.Value()
			fieldList := strings.Split(value, ",")
			for i := range fieldList {
				fieldList[i] = strings.TrimSpace(fieldList[i])
			}
			storedFields[name] = fieldList
		}
	}

	return storedFields
}

// Pull a field from the 'fields' map, using field mappings as available
func (c *IniFile) MapField(fields map[string]string, field string) (string, bool) {
	fieldMappings := c.Fields()
	fieldList, ok := fieldMappings[field]
	if ok {
		for _, f := range fieldList {
			value, ok := fields[f]
			if ok {
				return value, true
			}
		}
		return "", false
	} else {
		value, ok := fields[field]
		return value, ok
	}
}

// Reads the configuration file. The configuration is stored in a INI style file.
func readConfig(configPath string) (cfg *ini.File, err error) {
	configPath, err = filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("configuration file not found at %s", configPath)
	}

	if _, err2 := os.Stat(configPath); err2 != nil {
		return nil, fmt.Errorf("configuration file not found or not readable at %s", configPath)
	}

	cfg, err = ini.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("configuration file cannot be parsed at %s", configPath)
	}

	return cfg, nil
}
