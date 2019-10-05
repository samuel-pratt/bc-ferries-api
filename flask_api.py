## Made by Samuel Pratt
## Victoria, BC, Canada

from flask import Flask, jsonify
from ferry_times import get_data
from apscheduler.scheduler import Scheduler
import json

app = Flask(__name__)

sched = Scheduler() # Scheduler object
sched.start() # Start scheduler

@app.route('/')
def hello():
    return jsonify(get_data())
    #return jsonify(json.load(open('times.txt')))

##
# This if for if you read from a file that is scheduled to update every minute
# Currently it scrapes the site each time
##
def update():
    print('test')
    file = open('times.txt', 'w')
    with open('data.txt', 'w') as outfile:
        file.write(json.dumps(get_data()))
    file.close

#sched.add_interval_job(update, minutes=1)

if __name__ == '__main__':
    app.run()