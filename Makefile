code: 
	make -B proto && make -B server && make -B client && make -B userclient

proto: pb/configstore.proto
	cd pb/ && protoc -I . configstore.proto --go_out=plugins=grpc:. && cd ..

server: server/main.go
	cd server/ && go get . && go build && cd ..

client: client/main.go
	cd client/ && go get . && go build && cd ..

userclient: userclient/main.go
	cd userclient/ && go get . && go build && cd ..

docker:
	./create-docker-image.sh

dbsetup:
	kubectl create -f ./k8s-configs/mysql-pv.yml && kubectl create -f ./k8s-configs/mysql-deployment.yml

dblogin:
	kubectl run -it --rm --image=mysql:8.0 --restart=Never mysql-client -- mysql -h mysql -pdistributed_systems

db:
	kubectl create -f ./k8s-configs/mysql-deployment.yml

dbkill:
	kubectl delete deployment,svc mysql 
	
dbreset:
	kubectl delete deployment,svc mysql; kubectl delete pvc mysql-pv-claim && kubectl delete pv mysql-pv-volume

system:
	./launch-pods.sh

halt:
	kubectl delete services config-management-server configpollpythonweb configuration-update-client getconfigbash; kubectl delete ingresses.extensions configuration-management-ingress; kubectl delete deployments.apps configuration-update-client configpollpythonweb config-management-server getconfigbash

cleandocker:
	docker rm -f $(docker ps -a | grep -E '[a-z]+_[a-z]+$' | awk '{print $1}'); docker rmi $(docker images | grep '^<none>' | awk '{print $3}'); echo "Temporary docker images removed"

updateconfig:
	curl -X POST -L -k --header 'Content-Type: multipart/form-data' -F file=@"${CONFIG}" http://10.0.2.15:80/v1/update

clean:
	rm pb/configstore.pb.go && rm server/server && rm client/client

# docker run --name config_store -e MYSQL_ROOT_PASSWORD=distributed_system -e MYSQL_DATABASE=appconfig -d mysql:tag
# docker run --name some-app --link some-mysql:mysql -d application-that-uses-mysql