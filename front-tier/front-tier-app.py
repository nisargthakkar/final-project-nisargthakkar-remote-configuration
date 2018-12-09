import requests
from requests.auth import HTTPDigestAuth
import json
import time

# Replace with the correct URL
url = "http://localhost:8000/v1/config/k1"
while True:
	myResponse = requests.get(url)
	print(myResponse.status_code)
	if(myResponse.ok):
		print(myResponse.content)
	else:
		myResponse.raise_for_status()
	
	time.sleep(5)