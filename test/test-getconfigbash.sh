#!/bin/bash
cd /home/vagrant/go/src/github.com/nisargthakkar/final-project-nisargthakkar-remote-configuration

function updateconfig() {
	make updateconfig CONFIG=$1
}

function assert() {
	if [ "$1" != "$2" ]; then
		echo "Expected $1. Got $2"
		exit 1
	fi
}

function assertvalues() {
	assert $k1_val $1
	assert $k2_val $2
	assert $k3_val $3
	assert $k4_val $4
	echo "SUCCESS"
}

function checkapplogs() {
	appname=$1
	containernum=$2
	containername=$(kubectl get pods | grep $appname | head -n $containernum | tail -n 1 | awk '{print $1}')
	logout=$(kubectl logs $containername -c $appname | tail -n 20)

	k1_val=$(echo $logout | head -n 2 | tail -n 1 | grep -o \"value\":\"v1[^\"]*)
	k2_val=$(echo $logout | head -n 4 | tail -n 1 | grep -o \"value\":\"v2[^\"]*)
	k3_val=$(echo $logout | head -n 6 | tail -n 1 | grep -o \"value\":\"v3[^\"]*)
	k4_val=$(echo $logout | head -n 8 | tail -n 1 | grep -o \"value\":\"v4[^\"]*)
}

# v1, v2, v4
updateconfig "test/configs/getconfigbash_1.yml"
sleep 20

echo "Checking values on instance 1"
checkapplogs "getconfigbash" "1"
assertvalues "\"value\":\"v1" "\"value\":\"v2" "" "\"value\":\"v4"

echo "Checking values on instance 2"
checkapplogs "getconfigbash" "2"
assertvalues "\"value\":\"v1" "\"value\":\"v2" "" "\"value\":\"v4"

# v1_updated_again, v3, v4
updateconfig "test/configs/getconfigbash_2.yml"
sleep 20

echo "Checking values on instance 1"
checkapplogs "getconfigbash" "1"
assertvalues "\"value\":\"v1_updated_again" "" "\"value\":\"v3" "\"value\":\"v4"

echo "Checking values on instance 2"
checkapplogs "getconfigbash" "2"
assertvalues "\"value\":\"v1_updated_again" "" "\"value\":\"v3" "\"value\":\"v4"