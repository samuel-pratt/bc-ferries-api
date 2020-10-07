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

    # Routes
    route_links = {
        "Tsawwassen to Duke Point": "https://www.bcferries.com/current-conditions/vancouver-tsawwassen-nanaimo-duke-point/TSA-DUK",
        "Tsawwassen to Southern Gulf Islands": "https://www.bcferries.com/current-conditions/vancouver-tsawwassen-southern-gulf-islands/TSA-SGI",
        "Tsawwassen to Swartz Bay": "https://www.bcferries.com/current-conditions/vancouver-tsawwassen-victoria-swartz-bay/TSA-SWB",
        "Swartz Bay to Tsawwassen": "https://www.bcferries.com/current-conditions/victoria-swartz-bay-vancouver-tsawwassen/SWB-TSA",
        "Swartz Bay to Fulford Harbour": "https://www.bcferries.com/current-conditions/victoria-swartz-bay-salt-spring-island-fulford-harbour/SWB-FUL",
        "Swartz Bay to Southern Gulf Islands": "https://www.bcferries.com/current-conditions/victoria-swartz-bay-southern-gulf-islands/SWB-SGI",
        "Horseshoe Bay to Departure Bay": "https://www.bcferries.com/current-conditions/vancouver-horseshoe-bay-nanaimo-departure-bay/HSB-NAN",
        "Horseshoe Bay to Langdale": "https://www.bcferries.com/current-conditions/vancouver-horseshoe-bay-sunshine-coast-langdale/HSB-LNG",
        "Horseshoe Bay to Snug Cove": "https://www.bcferries.com/current-conditions/vancouver-horseshoe-bay-bowen-island-snug-cove/HSB-BOW",
        "Duke Point to Tsawwassen": "https://www.bcferries.com/current-conditions/nanaimo-duke-point-vancouver-tsawwassen/DUK-TSA",
        "Langdale to Horseshoe Bay": "https://www.bcferries.com/current-conditions/sunshine-coast-langdale-vancouver-horseshoe-bay/LNG-HSB",
        "Departure Bay to Horseshoe Bay": "https://www.bcferries.com/current-conditions/nanaimo-departure-bay-vancouver-horseshoe-bay/NAN-HSB"
    }

    # City - Terminal names
    ct_names = {
        "Tsawwassen": "Vancouver",
        "Horseshoe Bay": "Vancouver",
        "Swartz Bay": "Victoria",
        "Duke Point": "Nanaimo",
        "Departure Bay": "Nanaimo",
        "Langdale": "Sunshine Coast",
        "Snug Cove": "Bowen Island",
        "Fulford Harbour": "Salt Spring Island",
    }

    # Set webpage url and set up beautiful soup for scraping
    url = (route_links["Tsawwassen to Duke Point"])
    soup = make_soup(url)

    # Find all table data from webpage
    data = soup.find('table')
    df = pd.read_html(str(data))

    # Converts messy data into readable json
    raw_data = df[0].to_json(orient='records')
    json_data = json.loads(raw_data)[:-1]

    print(json_data)

    return times

# Used for testing
if __name__ == '__main__':
    print(get_data())