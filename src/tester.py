import requests

tUrl = "https:/"

def proxyTest():
    proxy = {
        "http":"http://localhost:8080"
    }
    r = requests.get(tUrl,proxies=proxy)
    requests.request()
    print(r.content)

proxyTest()