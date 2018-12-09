all: 
	make -B proto && make -B server && make -B client && make -B userclient

proto: pb/configstore.proto
	cd pb/ && protoc -I . configstore.proto --go_out=plugins=grpc:. && cd ..

server: server/main.go
	cd server/ && go get . && go build && cd ..

client: client/main.go
	cd client/ && go get . && go build && cd ..

userclient: userclient/main.go
	cd userclient/ && go get . && go build && cd ..

clean:
	rm pb/configstore.pb.go && rm server/server && rm client/client

# docker run --name config_store -e MYSQL_ROOT_PASSWORD=distributed_system -e MYSQL_DATABASE=appconfig -d mysql:tag
# docker run --name some-app --link some-mysql:mysql -d application-that-uses-mysql