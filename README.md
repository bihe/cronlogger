# Cronlogger

Capture the output of cron-tasks and store the in a sqlite DB.

- cmd/logger

Visualize the contents of the sqlite DB in a simple web UI.

- cmd/server

## Usage

### Logger
The cronlogger reads the piped output via stdin an stores it in the db

```bash
echo "${COMMAND_OUTPUT}" | /usr/local/bin/cronlogger \
    --app=<appname> \ 
    --code=${RESULT_CODE} \
    --db=/var/cronlog/cronlog-store.db
```

### Server
The server provides an http endpoint which shows the result of the executions. Typically a systemd service is used to start the server.

**example.service**:

```
[Unit]
Description=Cronlogger HTTP Server

[Service]
User=cronlogger
Group=cronlogger
Restart=always
ExecStart=/usr/local/bin/cronlogger_server -db /var/cronlog/cronlog-store.db -host localhost 

[Install]
WantedBy=multi-user.target
```
