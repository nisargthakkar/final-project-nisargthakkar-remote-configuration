import requests
from requests.auth import HTTPDigestAuth
import json
import time
import sys
import http.server
import socketserver
from threading import Thread
import os

MAX_CONNECTION_FAILURES = 5
remainingConnectionFailures = MAX_CONNECTION_FAILURES

def getConfig(key, defaultValue):
	global MAX_CONNECTION_FAILURES
	global remainingConnectionFailures
	url = "http://localhost:8000/v1/config/" + key
	print('Requesting config')
	try:
		myResponse = requests.get(url, timeout=0.1)
	except requests.exceptions.ReadTimeout as e:
		data = {}
		data[key] = defaultValue
		return data, defaultValue
	except requests.exceptions.ConnectionError as e:
		if remainingConnectionFailures == 0:
			print('Reached max failures. Exiting')
			os._exit(1)
		else:
			remainingConnectionFailures -= 1
			print('Connection failed. Using default value')
			data = {}
			data[key] = defaultValue
			return data, defaultValue
	
	remainingConnectionFailures = MAX_CONNECTION_FAILURES
	print(myResponse.status_code)
	if(myResponse.ok):
		data = myResponse.json()
		if 'value' in data:
			return data, data['value']
		
	print('Invalid response. Using default value')
	return data, defaultValue

def runBackground():
	LOG_TEXT = 'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Praesent feugiat massa tempor eros accumsan placerat. Maecenas non faucibus lorem. Suspendisse bibendum mollis risus, ac commodo dui venenatis ullamcorper. Donec aliquam nunc rhoncus viverra tempus. Aliquam auctor leo in neque viverra, sit amet scelerisque nibh commodo. Etiam lobortis porta erat vitae suscipit. Duis quis porttitor tellus, cursus lacinia orci. Proin posuere dapibus ex, nec commodo magna lacinia sit amet. Nam non consequat ligula. Vivamus non fringilla erat.\n\nProin eu risus eu quam fermentum malesuada. Sed ut neque tortor. Proin fringilla in ipsum ut lobortis. Praesent id dolor turpis. Sed sem enim, malesuada quis pretium vel, posuere vitae lacus. Quisque pharetra felis posuere dolor posuere elementum in ac nibh. Mauris at placerat lectus, in scelerisque erat. Fusce convallis luctus elit, non euismod enim condimentum pretium. Vestibulum eu leo hendrerit, vehicula magna et, luctus orci. In ac nunc sed sem malesuada luctus elementum in leo. Quisque feugiat ante nec hendrerit sagittis. Suspendisse rutrum enim sit amet eros volutpat feugiat.\n\nAenean euismod mi purus, eget laoreet est rutrum ut. Duis nibh lectus, tincidunt eget interdum nec, porta sit amet ligula. Pellentesque eget velit a erat viverra dignissim. Phasellus sagittis vehicula massa eget gravida. In non justo rutrum ligula imperdiet aliquet. Maecenas efficitur ipsum a sapien condimentum, lobortis dapibus dolor viverra. Donec sit amet ex eu ipsum aliquam fermentum. Fusce accumsan risus sed vestibulum eleifend. Etiam vitae quam nisl. Morbi fringilla ornare neque, ac condimentum nisl commodo in. Ut venenatis sem in massa dignissim vulputate. Vestibulum ac molestie diam. Integer mi purus, sodales ut pharetra in, sagittis ac est. Nulla at augue leo. Nulla facilisi. Mauris placerat risus id dolor porta ultrices.\n\nDonec leo nisl, consectetur et luctus id, suscipit sit amet lectus. Integer pharetra, velit non dignissim blandit, nisi odio tempor purus, nec hendrerit nisl ex vel orci. Nam a feugiat ipsum. Nam a nunc lacus. Aenean tincidunt sollicitudin aliquet. Nunc iaculis, nisl at fermentum venenatis, diam ligula mollis nulla, non sagittis urna enim eu urna. Curabitur sed odio odio.\n\nNunc aliquam tortor est, ac finibus felis pretium ut. Vestibulum lectus est, molestie in erat non, commodo iaculis leo. Phasellus sed arcu at leo aliquam auctor sit amet vel tellus. Nam scelerisque vehicula sapien, posuere volutpat lacus eleifend at. Fusce vestibulum sem id consequat porta. Integer non pretium elit, mollis viverra metus. Nulla facilisi. Nunc et volutpat mauris, nec molestie neque. Donec a felis et tellus iaculis scelerisque at id neque. Nam felis elit, placerat non dolor ut, sodales mattis nisl. Etiam massa lorem, imperdiet ultrices mi at, tincidunt eleifend lectus. Donec sed quam sapien. Quisque tempus mattis nulla eu gravida. Integer leo massa, volutpat eget odio quis, rhoncus hendrerit lacus. Curabitur eleifend dui vel metus rutrum, eu gravida ex porta.'
	while True:
		verboseLoggingData, verboseLogging = getConfig("verboseLogging", "False")
		print(verboseLoggingData)
		outfileData, filename = getConfig("outfile", "configout.txt")
		print(outfileData)
		file = open(filename, "w")
		print('verboseLogging:', verboseLogging == 'True')
		if verboseLogging == 'True':
			file.write(LOG_TEXT)
		else:
			file.write('VERBOSE_LOGGING_DISABLED')
		file.close()
		
		time.sleep(5)


t1 = Thread(target = runBackground)
t1.setDaemon(True)
t1.start()

# minimal web server.  serves files relative to the
# current directory.

PORT = 8080

Handler = http.server.SimpleHTTPRequestHandler

with socketserver.TCPServer(("", PORT), Handler) as httpd:
	print("serving at port", PORT)
	httpd.serve_forever()