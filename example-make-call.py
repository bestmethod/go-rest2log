import requests

# now carriage return
r = requests.post('http://127.0.0.1:5000/DEBUG',json={"Message":"Trying to do something funky and logging it"},headers={'Content-type': 'application/json', 'Accept': 'text/plain'})
print(r.status_code)
print(r.text)

# short request
r = requests.post('http://127.0.0.1:5000/DEBUG',json={"Message":"Trying to do something funky and logging it\r\nTesting carriage returns too."},headers={'Content-type': 'application/json', 'Accept': 'text/plain'})
print(r.status_code)
print(r.text)

# long request
longtext="1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678\n90123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"
r = requests.post('http://127.0.0.1:5000/DEBUG',json={"Message":longtext},headers={'Content-type': 'application/json', 'Accept': 'text/plain'})
print(r.status_code)
print(r.text)

