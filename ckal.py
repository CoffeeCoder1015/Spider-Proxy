import requests

sess = requests.session()
sess.get("http://localhost:8080",headers={"Connection":"keep-alive"})
sess.post("http://localhost:8080",data="HELLO WORLD")