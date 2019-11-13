# Doglog

Command-line interface to search and output logs from Datadog. Very useful for searching and tailing logs from the command-line. The default rate limiting for Datadog accounts and the Log Query API is 300 calls per hour. That is very, very low to use this utility. You will almost certainly need to request that limit to be raised.

The query syntax is defined here: https://docs.datadoghq.com/logs/explorer/search/#search-syntax

Originally came from https://github.com/bvargo/gtail. I converted it to Go and Datadog.

```text
usage: datadog [-h|--help] [-s|--service "<value>"] [-q|--query "<value>"]
               [-l|--limit <integer>] [-t|--tail] [-c|--config "<value>"]
               [-r|--range "<value>"] [--start "<value>"] [--end "<value>"]
               [-j|--json] [--no-colors]

               Search and tail logs from Datadog.

Arguments:

  -h  --help       Print help information
  -s  --service    Special case to search the 'service' message field, e.g., -s
                   send-email is equivalent to -q 'service:send-email'. Merged
                   with the -q query using 'AND' if the -q query is present.
  -q  --query      Query terms to search on (Doglog search syntax). Defaults to
                   '*'.
  -l  --limit      The maximum number of messages to request from Datadog. Must
                   be greater then 0. Default: 300
  -t  --tail       Whether to tail the output. Requires a relative search.
  -c  --config     Path to the config file. Default: /home/ctwise/.doglog
  -r  --range      Time range to search backwards from the current moment.
                   Examples: 30m, 2h, 4d. Default: 2h
      --start      Starting time to search from. Allows variable formats,
                   including '1:32pm' or '1/4/2019 12:30:00'.
      --end        Ending time to search from. Allows variable formats,
                   including '6:45am' or '2019-01-04 12:30:00'. Defaults to now
                   if --start is provided but no --end.
  -j  --json       Output messages in json format. Shows the modified log
                   message, not the untouched message from Datadog. Useful in
                   understanding the fields available when creating Format
                   templates or for further processing.
      --no-colors  Don't use colors in output.
```

Doglog requires a configuration file be setup in order to work. By default, the application looks in ~/.doglog.

A default configuration file might look like:

```ini
[server]
api-key: <API key>
application-key: <Application Key>

[fields]
; Define field mappings. You can leave this section out completely and it will use the below defaults.
; The mappings are for "special" fields and just include the below three.
; level: level, status, loglevel, log_status
; message: message, msg
; full_message: full_message, original_message
; classname: logger_name

[formats]
; log formats (list them most specific to least specific, they will be tried in order)
; all fields must be present or the format won't be applied
; Formats use the Go template syntax (https://golang.org/pkg/text/template/).

; Access logs (GET/POST, etc.)
; access log w/bytes
access_1: <{{.host}}> {{._long_time_timestamp}} {{._magenta}}{{.service}}{{._reset}} {{.network_client_ip}} {{.ident}} {{.auth}} "{{.http_method}} {{.http_url_details_path}} HTTP/{{.http_version}}" {{.http_status_code}} {{.network_bytes_read}}
; access log w/o bytes
access_2: <{{.host}}> {{._long_time_timestamp}} {._magenta}}{{.service}}{{._reset}} {{.network_client_ip}} {{.ident}} {{.auth}} "{{.http_method}} {{.http_url_details_path}} HTTP/{{.http_version}}" {{.http_status_code}}
; access log w/bytes
access_3: <{{.host}}> {{._long_time_timestamp}} {{._magenta}}{{.service}}{{._reset}} {{.network_client_ip}} "{{.http_method}} {{.http_url_details_path}} HTTP/{{.http_version}}" {{.http_status_code}} {{.network_bytes_read}}
; access log w/o bytes
access_4: <{{.host}}> {{._long_time_timestamp}} {{._magenta}}{{.service}}{{._reset}} {{.network_client_ip}} "{{.http_method}} {{.http_url_details_path}} HTTP/{{.http_version}}" {{.http_status_code}}
; access log w/bytes
access_5: <{{.host}}> {{._long_time_timestamp}} {{._magenta}}{{.service}}{{._reset}} {{.network_client_ip}} "{{.http_method}} {{.http_url_details_path}} HTTP/?" {{.http_status_code}} {{.network_bytes_read}}
; access log w/o bytes
access_6: <{{.host}}> {{._long_time_timestamp}} {{._magenta}}{{.service}}{{._reset}} {{.network_client_ip}} "{{.http_method}} {{.http_url_details_path}} HTTP/?" {{.http_status_code}}

; Java log entries (have thread and/or class names)
; java log entry 1
java_1: <{{.host}}> {{._long_time_timestamp}} {{._magenta}}{{.service}}{{._reset}} {{._level_color}}{{printf "%-5.5s" ._level}}{{._reset}} [{{printf "%-10.10s" .logger_thread_name}}] {{printf "%-20.20s" ._short_classname}} : {{._cyan}}{{._message_text}}{{._reset}}
; java log entry 2
java_2: <{{.host}}> {{._long_time_timestamp}} {{._magenta}}{{.service}}{{._reset}} {{._level_color}}{{printf "%-5.5s" ._level}}{{._reset}} {{printf "%-20.20s" ._short_classname}} : {{._cyan}}{{._message_text}}{{._reset}}
; java log entry 3
java_3: <{{.host}}> {{._long_time_timestamp}} {{._magenta}}{{.service}}{{._reset}} {{._level_color}}{{printf "%-5.5s" ._level}}{{._reset}} [{{printf "%-10.10s" .logger_thread_name}}] : {{._cyan}}{{._message_text}}{{._reset}}

; syslog
format8: <{{.host}}> {{._long_time_timestamp}} {{._magenta}}{{.service}}{{._reset}} {{._level_color}}{{printf "%-5.5s" ._level}}{{._reset}} [{{.syslog_appname}}] : {{._cyan}}{{._message_text}}{{._reset}}

; Istio mixer
; mixer _1
format9: <{{.host}}> {{._long_time_timestamp}} {{._magenta}}{{.service}}{{._reset}} {{._level_color}}{{printf "%-5.5s" ._level}}{{._reset}} {{.http_method}} {{.http_url_details_scheme}}:/{{.http_url_details_path}} {{.http_status_code}} {{.network_bytes_read}}
; mixer _2
format10: <{{.host}}> {{._long_time_timestamp}} {{._magenta}}{{.service}}{{._reset}} {{._level_color}}{{printf "%-5.5s" ._level}}{{._reset}} {{.http_url_details_scheme}} {{.totalSentBytes}} bytes -> {{.totalReceivedBytes}} bytes
; mixer _3
format11: <{{.host}}> {{._long_time_timestamp}} {{._magenta}}{{.service}}{{._reset}} {{._level_color}}{{printf "%-5.5s" ._level}}{{._reset}} {{.http_method}} {{.http_url_details_scheme}}:/{{.http_url_details_path}} {{.http_status_code}} {{.network_bytes_read}}

; vpc flow log
format_vpc_flow_log1: {{._long_time_timestamp}} {{._magenta}}{{.service}}{{._reset}} ({{.aws_account_id}}:{{printf "%-21.21s" .aws_s3_bucket}}) {{.vpc_action}} {{.network_client_ip}}:{{.network_client_port}} -> {{.network_destination_ip}}:{{.network_destination_port}} {{if eq (.network_protocol) ("6")}}TCP{{else if eq (.network_protocol) ("1")}}ICMP{{else if eq (.network_protocol) ("17")}}UDP{{else}}{{.network_protocol}}{{end}} {{.network_bytes_written | toString | trimSuffix ".0"}} bytes 
; vpc flow log no data
format_vpc_flow_log2: {{._long_time_timestamp}} {{._magenta}}{{.service}}{{._reset}} ({{.aws_account_id}}) : {{._cyan}}{{._message_text}}{{._reset}}

; generic
generic_1: <{{.host}}> {{._long_time_timestamp}} {{._magenta}}{{.service}}{{._reset}} {{._level_color}}{{printf "%-5.5s" ._level}}{{._reset}} : {{._cyan}}{{._message_text}}{{._reset}}
generic_2: <{{.host}}> {{._long_time_timestamp}} {{._magenta}}{{.service}}{{._reset}} : {{._cyan}}{{._message_text}}{{._reset}}
generic_3: {{._long_time_timestamp}} {{._magenta}}{{.service}}{{._reset}} : {{._cyan}}{{._message_text}}{{._reset}}
```

Multi-level field names have the period ('.') separator replaced by an underscore ('_'). For example, the multi-level field "network.protocol" is mapped to "network_protocol".

Fields that have special logic are level, message, full_message, classname. The default mappings for these special fields are:

| Field Name   | Mappings |
---------------|----------|
| level        | level, status, loglevel, log_status |
| message      | message, msg |
| full_message | full_message, original_message |
| classname    | logger_name |


Doglog creates some computed fields during log line processing. The computed fields are:

| Field Name          | Description |
|---------------------|-------------|
|_level               |The severity level of the log line, whether the incoming log has 'level', 'status', 'log_status' or 'loglevel', the 'level' field will be created and populated with a consistent severity.
|_level_color         |If the computed level field is generated, then an ASCII color code for the severity level will be present in this field.
|_reset               |Same as _level_color, but this resets the terminal color to Normal.
|_message_text        |The log line message text. Multiple fields are examined to generate this field.
|_long_time_timestamp |A consistent timestamp format for logging.
|_short_classname     |For Java log lines, this will be the short version of a full classname with package.

Template functions are provided from the Sprig template function library - http://masterminds.github.io/sprig/
