#!/usr/bin/env python3
import sys, os
from getopt import getopt
import json
import random
import psycopg2
import requests
from dotenv import load_dotenv
from haversine import inverse_haversine, Unit


API_URL_BASE = 'https://api.equinox.vision'
LAYER_ID = 93
N_IMAGES = 10

def main():
    load_dotenv()

    opts, args = getopt(sys.argv[1:], "h")
    for o, a in opts:
        if o == '-h':
            print("heh")
    if len(args) != 2:
        print("expecting arguments: coordinates lat, lng")
        return

    lat = float(args[0])
    lng = float(args[1])

    proposals_ori = json.load(open('media/proposals.json', 'rt'))
    proposals = {}
    for prop in proposals_ori:
        proposals[prop['id']] = prop

    cn = psycopg2.connect("dbname=sdplayground user=sdplayground")
    cur = cn.cursor()

    r = requests.post("%s/v1/login" % API_URL_BASE, json={"email": "ivoras@gmail.com", "password": os.environ['EQUINOX_PASSWORD'], "app_version": "1"})
    resp = r.json()
    if not resp['ok']: raise Exception("login failed")

    login_token = resp['login_token']

    objects = []
    cur.execute('select proposal_id,sum(grade) from grades group by proposal_id order by 2 desc limit 65')
    for row in cur:
        prop = proposals[row[0]]
        author_name = prop['author_name']
        if not 'ekipno' in author_name:
            parts = author_name.split(' ')
            author_name = "".join([x[0]+"." for x in parts if x != ""])
        obj = {
                "layer_id": LAYER_ID,
                "type": 46,
                "subtype": 1,
                "scale": 1,
                "message": "%s (%s): %s - %s" % (prop['author_about'], author_name, prop['artwork_title'], prop['artwork_description']),
                "label": prop['id'],
                "resource_url": "pending",
        }
        objects.append(obj)

    def post_obj(lat, lng, obj, image=None):
        if image:
            obj['resource_url'] = 'pending'
        obj['lat'] = lat
        obj['lng'] = lng
        r = requests.post("%s/v1/publish" % API_URL_BASE, json={"login_token": login_token, "object": obj})
        resp = r.json()
        if not resp['ok']:
            print("publishing error:")
            print(json)
            sys.exit(1)
        object_id = resp['object_id']
        if image:
            parts = image.split('.')
            extension = parts[-1]
            if extension == 'png':
                mime = 'image/png'
            elif extension in ('jpeg', 'jpg', 'jfif'):
                mime = 'image/jpeg'
            else:
                print("unknown file type: %s" % extension)
                sys.exit(1)
            headers = {'Content-type': mime}
            r = requests.put("%s/v1/attach?login_token=%s&object_id=%d" % (API_URL_BASE, login_token, object_id), data=open(image, 'rb'), headers=headers)
            resp = r.json()
            if not resp['ok']:
                print("cannot attach image file")
                print(resp)
                sys.exit(1)

    # Place the A logo in the center
    post_obj(lat, lng, {
        "layer_id": LAYER_ID,
        "type": 46,
        "subtype": 3,
        "scale": 1,
        "label": "A logo"})
        
    swing = 0 # in radians
    radius = 3
    for i in range(N_IMAGES):
        obj = random.choice(objects)
        prop = proposals[obj['label']]

        parts = prop['artwork_url'].split('.')
        extension = parts[-1]

        r = requests.get(prop['artwork_url'], auth=("artfuture", "ErRhBsGFG25w"))
        fname = "/tmp/image.%s" % extension
        with open(fname, 'wb') as f:
            f.write(r.content)

        new_lat, new_lng = inverse_haversine((lat, lng), radius, swing, unit=Unit.METERS)
        post_obj(new_lat, new_lng, obj, fname)

        radius += 1
        swing += 0.785398
        

if __name__ == '__main__':
    main()


