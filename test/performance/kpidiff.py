
import argparse
import sys
import yaml

ap = argparse.ArgumentParser()
ap.add_argument('-a', type=str, help='Base for comparison')
ap.add_argument('-b', type=str, help='Compared data')
ap.add_argument('-r', '--raw', action="store_true", default=False,
                help='Store differences as raw values (instead of percentages)')

args = ap.parse_args()

a_yaml_file = args.a
b_yaml_file = args.b
raw = bool(args.raw)

with open(a_yaml_file, "r") as a_yaml:
    with (open(b_yaml_file, "r") as b_yaml):
        a_kpi = yaml.safe_load(a_yaml)["kpi"]
        b_kpi = yaml.safe_load(b_yaml)["kpi"]

        kpi_i = 0
        for kpi in b_kpi:
            name = kpi["name"]
            metric_i = 0
            for metric in kpi["metrics"]:
                for stat in metric.keys():
                    if stat == "name":
                        continue
                    if a_kpi[kpi_i]["metrics"][metric_i][stat] == "n/a" or b_kpi[kpi_i]["metrics"][metric_i][stat] == "n/a":
                        value = "n/a"
                    else:
                        a = float(a_kpi[kpi_i]["metrics"][metric_i][stat])
                        b = float(b_kpi[kpi_i]["metrics"][metric_i][stat])
                        if a == 0:
                            if a != b:
                                c = float('inf')
                            else:
                                c = 0.0
                        else:
                            c = b / a - 1.0
                        if raw:
                            value = c
                        else:
                            if c > 0:
                                sign = "+"
                            else:
                                sign = ""
                            value = f"{sign}{round(c*100, 2)}%"
                    b_kpi[kpi_i]["metrics"][metric_i][stat] = value
                metric_i += 1
            kpi_i += 1
output = {"kpi-diff": b_kpi}

print(yaml.dump(output))
