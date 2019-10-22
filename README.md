# Doglog

Command-line interface to search and output logs from Datadog. Very useful for searching and tailing logs from the command-line. The default rate limiting for Datadog accounts and the Log Query API is 300 calls per hour. That is very, very low to use this utility. You will almost certainly need to request that limit to be raised.

The query syntax is defined here: https://docs.datadoghq.com/logs/explorer/search/#search-syntax

Originally came from https://github.com/bvargo/gtail. I converted it to Go and Datadog.

```text
usage: datadog [-h|--help] [-a|--application "<value>"] [-q|--query "<value>"]
               [-l|--limit <integer>] [-t|--tail] [-c|--config "<value>"]
               [-r|--range "<value>"] [--start "<value>"] [--end "<value>"]
               [-j|--json] [--no-colors]

               Search and tail logs from Datadog.

Arguments:

  -h  --help         Print help information
  -a  --application  Special case to search the 'application' message field,
                     e.g., -a send-email is equivalent to -q
                     'application:send-email'. Merged with the -q query using
                     'AND' if the -q query is present.
  -q  --query        Query terms to search on (Elasticsearch syntax). Defaults
                     to '*'.
  -l  --limit        The maximum number of messages to request from Datadog.
                     Must be greater then 0. Default: 300
  -t  --tail         Whether to tail the output. Requires a relative search.
  -c  --config       Path to the config file. Default: /home/ctwise/.doglog
  -r  --range        Time range to search backwards from the current moment.
                     Examples: 30m, 2h, 4d. Default: 2h
      --start        Starting time to search from. Allows variable formats,
                     including '1:32pm' or '1/4/2019 12:30:00'.
      --end          Ending time to search from. Allows variable formats,
                     including '6:45am' or '2019-01-04 12:30:00'. Defaults to
                     now if --start is provided but no --end.
  -j  --json         Output messages in json format. Shows the modified log
                     message, not the untouched message from Datadog. Useful in
                     understanding the fields available when creating Format
                     templates or for further processing.
      --no-colors    Don't use colors in output.
```

Doglog requires a configuration file be setup in order to work. By default, the application looks in ~/.doglog.

A default configuration file might look like:

```ini
[server]
api-key: <API key>
application-key: <Application Key>
[formats]
; log formats (list them most specific to least specific, they will be tried in order)
; all fields must be present or the format won't be applied
; Formats use the Go template syntax.
;
; access log w/bytes
format1: <{{.host}}> {{.client_ip}} {{.ident}} {{.auth}} [{{.apache_timestamp}}] "{{.method}} {{.request_page}} HTTP/{{.http_version}}" {{.server_response}} {{.bytes}}
; access log w/o bytes
format2: <{{.host}}> {{.client_ip}} {{.ident}} {{.auth}} [{{.apache_timestamp}}] "{{.method}} {{.request_page}} HTTP/{{.http_version}}" {{.server_response}}
; java log entry
format3: <{{.host}}> {{._long_time_timestamp}} {{._level_color}}{{printf "%-5.5s" .loglevel}}{{._reset}} {{printf "%-20.20s" ._short_classname}} : {{._message_text}}
; syslog
format4: <{{.host}}> {{._long_time_timestamp}} {{._level_color}}{{printf "%-5.5s" .loglevel}}{{._reset}} [{{.facility}}] : {{._message_text}}
; generic entry with a loglevel
format5: <{{.host}}> {{._long_time_timestamp}} {{._level_color}}{{printf "%-5.5s" .loglevel}}{{._reset}} : {{._message_text}}
```

Doglog creates some computed fields during log line processing. The computed fields are:

loglevel - The severity level of the log line, whether the incoming log has 'level', 'status', 'log_status' or 'loglevel', the 'loglevel' field will be created and populated with a consistent severity.
_level_color - If the computed loglevel field is generated, then an ASCII color code for the severity level will be present in this field.
_reset - Same as _level_color, but this resets the terminal color to Normal.
_message_text - The log line message text. Multiple fields are examined to generate this field.
_long_time_timestamp - A consistent timestamp format for logging.
_short_classname - For Java log lines, this will be the short version of a full classname with package.
