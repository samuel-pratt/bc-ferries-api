import os
import json
from flask_cors import CORS
from flask import Flask, jsonify, render_template
from flata import Flata, where
from flata.storages import JSONStorage
from apscheduler.scheduler import Scheduler
from scraper import get_data

# Initialize db
db = Flata('./ferries.json', storage=JSONStorage)
db.table('table1')
db.get('table1')

app = Flask(__name__)
CORS(app)

# Scheduler object
sched = Scheduler(daemon=True)
sched.start()

# Valid departure terminals
departure_terminals = [
    'tsawwassen',
    'swartz-bay',
    'nanaimo-(duke-pt)',
    'nanaimo-(dep.bay)',
    'horseshoe-bay',
    'langdale',
    'snug-cove-(bowen-is.)'
]

# Valid destination terminals
destination_terminals = {
    'tsawwassen': [
        'swartz-bay',
        'southern-gulf-islands',
        'nanaimo-(duke-pt)'
    ],
    'swartz-bay': [
        'tsawwassen',
        'fulford-harbour',
    ],
    'nanaimo-(duke-pt)': [
        'tsawwassen'
    ],
    'nanaimo-(dep.bay)': [
        'horseshoe-bay'
    ],
    'horseshoe-bay': [
        'nanaimo-(dep.bay)',
        'langdale',
        'snug-cove-(bowen-is.)'
    ],
    'langdale': [
        'horseshoe-bay'
    ],
    'snug-cove-(bowen-is.)': [
        'horseshoe-bay'
    ]
}

# Update data every 3 minutes
@sched.interval_schedule(minutes=3)
def updateDb():
    with app.app_context():
        results = get_data()
        db.table('table1').update({'data': results}, where('id') ==1)
        print('Updated ferries.json')

@app.route('/')
def home_page():
    return render_template('index.html')

# Returns all route data
@app.route('/api/')
def all_data():
    data = db.all()
    return jsonify(data['table1'][0]['data'])

# Returns all info for a specified terminal
@app.route('/api/<departure_terminal>/')
def terminal_data(departure_terminal):
    departure_terminal = departure_terminal.lower()

    # Check that terminal is valid
    if departure_terminal not in departure_terminals:
        return jsonify('Error: not a valid departure terminal.')

    # Format paramater for accessing db
    departure_terminal = departure_terminal.replace('-', ' ')

    data = db.all()
    return jsonify(data['table1'][0]['data'][departure_terminal])

# Returns all info for a specified route
@app.route('/api/<departure_terminal>/<destination_terminal>/')
def terminals_data(departure_terminal, destination_terminal):
    departure_terminal = departure_terminal.lower()
    destination_terminal = destination_terminal.lower()

    # Check that departure terminal is valid
    if departure_terminal not in departure_terminals:
        return jsonify('Error: Not a valid departure terminal.')

    # Cheheck that destination terminal is valid
    if destination_terminal not in destination_terminals[departure_terminal]:
        return jsonify('Error: Not a valid destination terminal.')

    # Format paramaters for accessing db
    departure_terminal = departure_terminal.replace('-', ' ')
    destination_terminal = destination_terminal.replace('-', ' ')

    data = db.all()
    return jsonify(data['table1'][0]['data'][departure_terminal][destination_terminal])

if __name__ == '__main__':
    app.run(debug=True,host='0.0.0.0',port=int(os.environ.get('PORT', 8080)))
