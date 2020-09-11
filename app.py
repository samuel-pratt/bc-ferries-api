## Made by Samuel Pratt
## Victoria, BC, Canada

import os

from flask import Flask, jsonify, render_template
from flata import Flata, where
from flata.storages import JSONStorage
from apscheduler.scheduler import Scheduler

from scraper import get_data

import json

db = Flata('./db.json', storage=JSONStorage)
db.table('table1')
db.get('table1')

app = Flask(__name__)

sched = Scheduler(daemon=True) # Scheduler object
sched.start()

# Valid terminals
departure_terminals = [
    "tsawwassen",
    "swartz-bay",
    "nanaimo-(duke-pt)",
    "nanaimo-(dep.bay)",
    "horseshoe-bay",
    "langdale"
]

destination_terminals = {
    "tsawwassen": [
        "swartz-bay",
        "southern-gulf-islands",
        "duke-point"
    ],
    "swartz-bay": [
        "tsawwassen",
        "fulford-harbour",
    ],
    "nanaimo-(duke-pt)": [
        "tsawwassen"
    ],
    "nanaimo-(dep.bay)": [
        "horseshoe-bay"
    ],
    "horseshoe-bay": [
        "departure-bay",
        "langdale",
        "snug-cove-bowen-island"
    ],
    "langdale": [
        "horseshoe-bay"
    ]
}

data_types = [
    "next-sailings",
    "future-sailings",
    "car-waits",
    "oversize-waits"
]

# Update data every minute
@sched.interval_schedule(minutes=1)
def updateDb():
    with app.app_context():
        results = get_data()
        db.table('table1').update({'data': results}, where('id') ==1)
        print('Updated DB')

@app.route('/')
def home_page():
    return render_template("index.html")

# Returns all route data
@app.route('/api/')
def all_data():
    data = db.all()
    return jsonify(data['table1'][0]['data'])

# Return all routes for a specified terminal
@app.route('/api/<departure_terminal>/')
def terminal_data(departure_terminal):
    # Check that terminal is valid
    if departure_terminal not in departure_terminals:
        return jsonify("Error: not a valid terminal.")

    departure_terminal = departure_terminal.replace('-', ' ')

    data = db.all()
    return jsonify(data['table1'][0]['data'][departure_terminal])

# Return all info for 
@app.route('/api/<departure_terminal>/<destination_terminal>/')
def terminals_data(departure_terminal, destination_terminal):
    # Check that departure terminal is valid
    if departure_terminal not in departure_terminals:
        return jsonify("Error: Not a valid departure terminal.")

    # Cheheck that destination terminal is valid
    if destination_terminal not in destination_terminals[departure_terminal]:
        return jsonify("Error: Not a valid destination terminal.")

    departure_terminal = departure_terminal.replace('-', ' ')
    destination_terminal = destination_terminal.replace('-', ' ')

    data = db.all()
    return jsonify(data['table1'][0]['data'][departure_terminal][destination_terminal])

@app.route('/api/<departure_terminal>/<destination_terminal>/<data_type>/')
def info_data(departure_terminal, destination_terminal, data_type):
    # Check that departure terminal is valid
    if departure_terminal not in departure_terminals:
        return jsonify("Error: Not a valid departure terminal.")

    # Check that destination terminal is valid
    if destination_terminal not in destination_terminals[departure_terminal]:
        return jsonify("Error: Not a valid destination terminal.")

    # Check that data type is valid
    if data_type not in data_types:
        return jsonify("Error: Not a valid data type.")

    departure_terminal = departure_terminal.replace('-', ' ')
    destination_terminal = destination_terminal.replace('-', ' ')
    data_type = data_type.replace('-', ' ')

    data = db.all()
    return jsonify(data['table1'][0]['data'][departure_terminal][destination_terminal][data_type])

if __name__ == '__main__':
    app.run(debug=True,host='0.0.0.0',port=int(os.environ.get('PORT', 8080)))