HOST=46.202.143.194
HOMEDIR=/var/www/user-service/
USER=dima

user-service-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/user-service-linux-amd64 ./

upload-user-service: user-service-linux
	rsync -rzv --progress --rsync-path="sudo rsync" \
		./bin/user-service-linux-amd64  \
		./utils/cfg/prod.ini \
		./utils/restart.sh \
		$(USER)@$(HOST):$(HOMEDIR)

restart-user-service:
	echo "sudo su && cd $(HOMEDIR) && bash restart.sh && exit" | ssh $(USER)@$(HOST) /bin/sh

upload-and-restart: upload-user-service restart-user-service

run-local:
	go run main.go -config ./utils/cfg/local.ini