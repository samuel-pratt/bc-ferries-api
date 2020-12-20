## Made by Samuel Pratt
## Victoria, BC, Canada

import json
import requests
import pandas as pd
from bs4 import BeautifulSoup
import pprint

# Took this off of beautiful soup documentation
def make_soup(url):
    res = requests.get(url, verify=False)
    html_content = res.text
    soup_object = BeautifulSoup(html_content , "html.parser")
    return soup_object

def get_data():
    # Data layout for schedule
    schedule = {
        "tsawwassen": {
            "swartz bay": [],
            "southern gulf islands": [],
            "nanaimo (duke pt)": []
        },
        "swartz bay": {
            "tsawwassen": [],
            "fulford harbour (saltspring is.)": [],
            "southern gulf islands": []
        },
        "nanaimo (duke pt)": {
            "tsawwassen": []
        },
        "nanaimo (dep.bay)": {
            "horseshoe bay": []
        },
        "horseshoe bay": {
            "nanaimo (dep.bay)": [],
            "langdale": [],
            "snug cove (bowen is.)": []
        },
        "langdale": {
            "horseshoe bay": []
        },
        "snug cove (bowen is.)": {
            "horseshoe bay": [],
        }
    }

    # BC Ferries link for each route
    route_links = {
        "tsawwassen to nanaimo (duke pt)": "https://www.bcferries.com/current-conditions/vancouver-tsawwassen-nanaimo-duke-point/TSA-DUK",
        "tsawwassen to southern gulf islands": "https://www.bcferries.com/current-conditions/vancouver-tsawwassen-southern-gulf-islands/TSA-SGI",
        "tsawwassen to swartz bay": "https://www.bcferries.com/current-conditions/vancouver-tsawwassen-victoria-swartz-bay/TSA-SWB",
        "swartz bay to tsawwassen": "https://www.bcferries.com/current-conditions/victoria-swartz-bay-vancouver-tsawwassen/SWB-TSA",
        "swartz bay to fulford harbour (saltspring is.)": "https://www.bcferries.com/current-conditions/victoria-swartz-bay-salt-spring-island-fulford-harbour/SWB-FUL",
        "swartz bay to southern gulf islands": "https://www.bcferries.com/current-conditions/victoria-swartz-bay-southern-gulf-islands/SWB-SGI",
        "horseshoe bay to nanaimo (dep.bay)": "https://www.bcferries.com/current-conditions/vancouver-horseshoe-bay-nanaimo-departure-bay/HSB-NAN",
        "horseshoe bay to langdale": "https://www.bcferries.com/current-conditions/vancouver-horseshoe-bay-sunshine-coast-langdale/HSB-LNG",
        "horseshoe bay to snug cove (bowen is.)": "https://www.bcferries.com/current-conditions/vancouver-horseshoe-bay-bowen-island-snug-cove/HSB-BOW",
        "nanaimo (duke pt) to tsawwassen": "https://www.bcferries.com/current-conditions/nanaimo-duke-point-vancouver-tsawwassen/DUK-TSA",
        "langdale to horseshoe bay": "https://www.bcferries.com/current-conditions/sunshine-coast-langdale-vancouver-horseshoe-bay/LNG-HSB",
        "nanaimo (dep.bay) to horseshoe bay": "https://www.bcferries.com/current-conditions/nanaimo-departure-bay-vancouver-horseshoe-bay/NAN-HSB",
        "snug cove (bowen is.) to horseshoe bay": "https://www.bcferries.com/routes-fares/schedules/-/BOW-HSB"
    }

    # Route names
    routes = [
        "tsawwassen to nanaimo (duke pt)",
        "tsawwassen to southern gulf islands",
        "tsawwassen to swartz bay",
        "swartz bay to tsawwassen",
        "swartz bay to fulford harbour (saltspring is.)",
        "swartz bay to southern gulf islands",
        "horseshoe bay to nanaimo (dep.bay)",
        "horseshoe bay to langdale",
        "horseshoe bay to snug cove (bowen is.)",
        "nanaimo (duke pt) to tsawwassen",
        "langdale to horseshoe bay",
        "nanaimo (dep.bay) to horseshoe bay",
        "snug cove (bowen is.) to horseshoe bay"
    ]

    for route in routes:
        # Set webpage url and set up beautiful soup for scraping
        url = (route_links[route])
        soup = make_soup(url)

        # Find all table data from webpage
        data = soup.find('table')
        df = pd.read_html(str(data))

        # Converts messy data into readable json
        raw_data = df[0].to_json(orient='records')
        json_data = json.loads(raw_data)[:-1]

        index = route.split(' to ')

        for i in json_data[:]:
            if 'Depart' in i.keys():
                sailing_data = {
                     "time": i['Depart'],
                     "capacity": "Unknown"
                }

            else:
                if len(i['1']) >= 20:
                    continue
                
                if 'Status' in i['1'] or 'Arrived' in i['1'] or 'ETA' in i['1']:
                    continue

                sailing_data = {
                   "time": i['0'],
                   "capacity": i['1'].split(' ')[0]
                }
            
            schedule[index[0]][index[1]].append(sailing_data)

    return schedule

# Used for testing
if __name__ == '__main__':
    print( get_data() )