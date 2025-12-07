# Cronlogger

Capture the output of cron-tasks and store the output in a sqlite DB.

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

**cronlogger_server.service**:

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

## Deployment
A simple git-deployment was established for the target-system following ths gist: https://gist.github.com/noelboss/3fe13927025b89757f8fb12e9066f2fa

```bash
# steps performed on the target-system

# the target folder to clone artifacts to
mkdir ~/cronlogger.deployment

# create a deployment git-repo
git init --bare ~/cronlogger.git

# create a post-receive hook script
touch cronlogger.git/hooks/post-receive
chmod +x post-receive

# -----------------------------------------------------------------------------------
# local/development system
# add remote on local/development system
git remote add production user@system:cronlogger.git
```

The post-receive hook takes care of the deployment

**post-receive**:
```bash
#!/bin/bash
TARGET_FOLDER="/path/to/cronlogger.deployment"
DEPLOYMENT_FOLDER="/usr/local/bin"
CONFIG_FOLDER="/etc/cronlogger"
GIT_DIR="/path/to/cronlogger.git"
BRANCH="main"

while read oldrev newrev ref
do
        # only checking out the master (or whatever branch you would like to deploy)
        if [ "$ref" = "refs/heads/$BRANCH" ];
        then
                echo "Ref $ref received. Deploying ${BRANCH} branch to production..."
                git --work-tree=$TARGET_FOLDER --git-dir=$GIT_DIR checkout -f $BRANCH
                sudo cp -f ${TARGET_FOLDER}/cronlogger ${DEPLOYMENT_FOLDER}/cronlogger
                sudo cp -f ${TARGET_FOLDER}/cronlogger_server ${DEPLOYMENT_FOLDER}/cronlogger_server
                sudo cp -f ${TARGET_FOLDER}/application.yaml ${CONFIG_FOLDER}/application.yaml
                sudo systemctl restart cronlogger_server
                sudo systemctl status cronlogger_server
                echo "Deployment done; server restarted; have fun!"
        else
                echo "Ref $ref received. Doing nothing: only the ${BRANCH} branch may be deployed on this server."
        fi
done


```
