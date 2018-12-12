#!/bin/bash
cd /home/vagrant/go/src/github.com/nisargthakkar/final-project-nisargthakkar-remote-configuration

function updateconfig() {
	make updateconfig CONFIG=$1
}

function getfilewordcount() {
	numwords=$(curl -L -k http://10.0.2.15:80/$1 | wc --words)
	if [[ $numwords -ne $2 ]]; then
		echo "Expected $2 words. Got $numwords words"
		echo "CALL FAILED"
		exit 1
	else
		echo "SUCCESS"
	fi
}

function getfilewordcount_parallel() {
	for i in `seq 1 $1`;
	do
		getfilewordcount $2 $3 &
	done
	wait
}

# outfile1.txt True
updateconfig "test/configs/configpollpythonweb_1.yml"
sleep 20

getfilewordcount_parallel "5" "outfile1.txt" "457"

# outfile2.txt False
updateconfig "test/configs/configpollpythonweb_2.yml"
sleep 20

getfilewordcount_parallel "5" "outfile2.txt" "1"
getfilewordcount_parallel "5" "outfile2.txt" "1"

# outfile2.txt True
updateconfig "test/configs/configpollpythonweb_3.yml"
sleep 20

getfilewordcount_parallel "5" "outfile2.txt" "457"
getfilewordcount_parallel "5" "outfile2.txt" "457"