## Made by Samuel Pratt
## Victoria, BC, Canada

from flask import Flask, jsonify
from flata import Flata, where
from flata.storages import JSONStorage
from apscheduler.scheduler import Scheduler

from scraper import get_data

import json

db = Flata('./db.json', storage=JSONStorage)
db.table('table1')
db.get('table1')

app = Flask(__name__)

sched = Scheduler() # Scheduler object
sched.start()

@sched.interval_schedule(minutes=1)
def updateDb():
    with app.app_context():
        results = get_data()
        db.table('table1').update({'data': results}, where('id') ==1)
        print('Updated DB')


@app.route('/')
def hello():
    temp = db.all()
    return temp['table1'][0]['data']

if __name__ == '__main__':
    app.run()