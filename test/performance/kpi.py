
import argparse
import csv
import yaml
import sys
import statistics
from datetime import datetime


def normalize_value(value):
    if float(value) < 0:
        return "n/a"
    else:
        return value


ap = argparse.ArgumentParser()
ap.add_argument('-c', '--csv-file', type=str, help='')
ap.add_argument('-x', '--x-column', type=int, help='')
ap.add_argument('-y', '--y-column', type=int, help='')
ap.add_argument('-d', '--date-format', type=str,
                default='%Y-%m-%d %H:%M:%S.%f', help='')

args = ap.parse_args()

csv_file = args.csv_file
x_column = args.x_column
y_column = args.y_column
dateformat = args.date_format

with open(csv_file, "r") as file:
    csvreader = csv.reader(file, delimiter=";")

    headers = []
    headers = next(csvreader)

    x_header = headers[x_column]
    y_header = headers[y_column]

    rows = []
    for row in csvreader:
        if row[x_column] == "" or row[y_column] == "" or row[x_column] is None or row[y_column] is None:
            sys.stderr.write(
                f"WARNING: Incomplete row data {row}, skipping...\n")
            continue
        rowmap = {}
        if len(row[x_column]) > 26:
            x = row[x_column][0:26]
        else:
            x = row[x_column]
        rowmap[x_header] = datetime.strptime(x, dateformat).timestamp()
        rowmap[y_header] = row[y_column]
        rows.append(rowmap)

count = len(rows)

min = float("inf")
max = float("-inf")
values = []

for rowmap in rows:
    value = float(rowmap[y_header])
    values.append(value)

    if value > max:
        max = value

    if value < min:
        min = value

metrics = {}
metrics["name"] = y_header
if count > 0:
    metrics["first"] = normalize_value(float(rows[0][y_header]))
    metrics["minimum"] = normalize_value(min)
    metrics["average"] = normalize_value(statistics.mean(values))
    metrics["median"] = normalize_value(statistics.median(values))
    metrics["maximum"] = normalize_value(max)
    metrics["last"] = normalize_value(float(rows[count - 1][y_header]))
else:
    metrics["first"] = "n/a"
    metrics["minimum"] = "n/a"
    metrics["average"] = "n/a"
    metrics["median"] = "n/a"
    metrics["maximum"] = "n/a"
    metrics["last"] = "n/a"

print(yaml.dump([metrics]))
