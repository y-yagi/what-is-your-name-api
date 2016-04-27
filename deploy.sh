#!/bin/sh
cp -p app.yaml app.yaml.back
cp -p app.yaml.production app.yaml
goapp deploy -application $1 app.yaml
mv app.yaml.back app.yaml
