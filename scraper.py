## Made by Samuel Pratt
## Victoria, BC, Canada

import json
import pandas as pd
import requests
from bs4 import BeautifulSoup

# Took this off of beautiful soup documentation
# add url here
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
        "Tsawwassen": {
            "Swartz Bay": {
                "Next Sailings": [],
                "Future Sailings": [],
                "Car Waits": 0,
                "Oversize Waits": 0
            },
            "Southern Gulf Islands": {
                "Next Sailings": [],
                "Future Sailings": [],
                "Car Waits": 0,
                "Oversize Waits": 0
            },
            "Duke Point": {
                "Next Sailings": [],
                "Future Sailings": [],
                "Car Waits": 0,
                "Oversize Waits": 0
            }
        },
        "Swartz Bay": {
            "Tsawwassen": {
                "Next Sailings": [],
                "Future Sailings": [],
                "Car Waits": 0,
                "Oversize Waits": 0
            },
            "Fulford Harbour (Saltspring Is.)": {
                "Next Sailings": [],
                "Future Sailings": [],
                "Car Waits": 0,
                "Oversize Waits": 0
            },
            "Southern Gulf Islands": {
                "Next Sailings": [],
                "Future Sailings": [],
                "Car Waits": 0,
                "Oversize Waits": 0
            }
        },
        "Nanaimo (Duke Pt)": {
            "Tsawwassen": {
                "Next Sailings": [],
                "Future Sailings": [],
                "Car Waits": 0,
                "Oversize Waits": 0
            }
        },
        "Nanaimo (Dep.Bay)": {
            "Horseshoe Bay": {
                "Next Sailings": [],
                "Future Sailings": [],
                "Car Waits": 0,
                "Oversize Waits": 0
            }
        },
        "Horseshoe Bay": {
            "Departure Bay": {
                "Next Sailings": [],
                "Future Sailings": [],
                "Car Waits": 0,
                "Oversize Waits": 0
            },
            "Langdale": {
                "Next Sailings": [],
                "Future Sailings": [],
                "Car Waits": 0,
                "Oversize Waits": 0
            },
            "Snug Cove (Bowen Is.)": {
                "Next Sailings": [],
                "Future Sailings": [],
                "Car Waits": 0,
                "Oversize Waits": 0
            }
        },
        "Langdale": {
            "Horseshoe Bay": {
                "Next Sailings": [],
                "Future Sailings": [],
                "Car Waits": 0,
                "Oversize Waits": 0
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
        if len(x) == 0:
            continue
        elif len(x) == 1:
            terminal = x['0']
        elif '4' in x.keys():
            destination = x['0'].split(' to ')[1]
            if '2' in x.keys():
                times[terminal][destination]['Car Waits'] = x['2']
            if '3' in x.keys():
                times[terminal][destination]['Oversize Waits'] = x['3']
            times[terminal][destination]['Future Sailings'] = x['4'].split()
        elif '1' in x.keys():
            times[terminal][destination]['Next Sailings'].append([x['0'],x['1']])

    return times

# Used for testing
if __name__ == '__main__':
    print(get_data())