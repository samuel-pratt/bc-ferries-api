## Made by Samuel Pratt
## Victoria, BC, Canada

import json
import requests
import pandas as pd
from bs4 import BeautifulSoup

# Took this off of beautiful soup documentation
def make_soup(url):
    res = requests.get(url)
    html_content = res.text
    soup_object = BeautifulSoup(html_content , "html.parser")
    return soup_object

def get_data():
    # Set webpage url and set up beautiful soup for scraping
    url = ("https://orca.bcferries.com/cc/marqui/at-a-glance.asp")
    soup = make_soup(url)

    # Find all table data from webpage
    data = soup.find('table')
    df = pd.read_html(str(data))

    # Data layout for schedule
    times = {
        "tsawwassen": {
            "swartz bay": {
                "next sailings": [],
                "future sailings": [],
                "car waits": 0,
                "oversize waits": 0
            },
            "southern gulf islands": {
                "next sailings": [],
                "future sailings": [],
                "car waits": 0,
                "oversize waits": 0
            },
            "duke point": {
                "next sailings": [],
                "future sailings": [],
                "car waits": 0,
                "oversize waits": 0
            }
        },
        "swartz bay": {
            "tsawwassen": {
                "next sailings": [],
                "future sailings": [],
                "car waits": 0,
                "oversize waits": 0
            },
            "fulford harbour (saltspring is.)": {
                "next sailings": [],
                "future sailings": [],
                "car waits": 0,
                "oversize waits": 0
            },
            "southern gulf islands": {
                "next sailings": [],
                "future sailings": [],
                "car waits": 0,
                "oversize waits": 0
            }
        },
        "nanaimo (duke pt)": {
            "tsawwassen": {
                "next sailings": [],
                "future sailings": [],
                "car waits": 0,
                "oversize waits": 0
            }
        },
        "nanaimo (dep.bay)": {
            "horseshoe bay": {
                "next sailings": [],
                "future sailings": [],
                "car waits": 0,
                "oversize waits": 0
            }
        },
        "horseshoe bay": {
            "departure bay": {
                "next sailings": [],
                "future sailings": [],
                "car waits": 0,
                "oversize waits": 0
            },
            "langdale": {
                "next sailings": [],
                "future sailings": [],
                "car waits": 0,
                "oversize waits": 0
            },
            "snug cove (bowen is.)": {
                "next sailings": [],
                "future sailings": [],
                "car waits": 0,
                "oversize waits": 0
            }
        },
        "langdale": {
            "horseshoe bay": {
                "next sailings": [],
                "future sailings": [],
                "car waits": 0,
                "oversize waits": 0
            }
        }
    }

    # Converts messy data into readable json
    raw_data = df[0].to_json(orient='records')
    json_data = json.loads(raw_data)[:-1]

    # Cleans out useless data from json_data
    for x in json_data:
        for y in range(8):
            if x[str(y)] == None or 'Conditions' in x[str(y)] or '*denotes' in x[str(y)] or 'Route' in x[str(y)] or 'Next Sailings' in x[str(y)] or 'Car Waits' in x[str(y)] or 'Oversize Waits' in x[str(y)] or 'Later Sailings' in x[str(y)] or 'Depart, Arrive' in x[str(y)] or 'Service Notices' in x[str(y)] or 'Sailing Details' in x[str(y)]:
                x.pop(str(y))

    # Takes out useful data from json_data and puts it into times
    for x in json_data:
        try:
            if len(x) == 0:
                continue
            elif len(x) == 1:
                terminal = x['0'].lower()
            elif '4' in x.keys():
                destination = x['0'].split(' to ')[1].lower()
                if '2' in x.keys():
                    times[terminal][destination]['car waits'] = x['2']
                if '3' in x.keys():
                    times[terminal][destination]['oversize waits'] = x['3']
                future_sailings = x['4'].split()
                for i in range(len(future_sailings)):
                    future_sailings[i] = future_sailings[i].replace('*', '')
                times[terminal][destination]['future sailings'] = future_sailings
            elif '1' in x.keys():
                times[terminal][destination]['next sailings'].append([x['0'],x['1']])
        except KeyError:
            # If there is a warning on BC Ferries' site, this will catch it
            # The data should still be accurately updated
            print("KeyError")
            continue

    return times

# Used for testing
if __name__ == '__main__':
    print(get_data())