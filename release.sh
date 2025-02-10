#!/bin/sh

release=$1

sed -i "s/tag: .*$/tag: $release/"  charts/kubelish/values.yaml
sed -i "s/appVersion: .*$/appVersion: $release/"  charts/kubelish/Chart.yaml
sed -i "s/version: .*$/version: $release/"  charts/kubelish/Chart.yaml
