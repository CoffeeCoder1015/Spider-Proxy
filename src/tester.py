import requests
import datetime
import time

resources = ["/", "/stuff.js", "/CloseButton.png",
             "/100", "/math?eq=20%2b23%2b4^(3/2)", "/math.ans"]


def unitTest() -> float:
    Start = time.time()
    for r in resources:
        url = f"https://localhost:8080{r}"
        print(r, url)
        r = requests.get(url)
        print(r)
    return round(time.time()-Start,8)

print(unitTest(),"secs")

