import os
from string import Template


def scenario_id(context):
    return f"{os.path.basename(os.path.splitext(context.scenario.filename)[0]).lower()}-{context.scenario.line}"


def substitute_scenario_id(context, text="$scenario_id"):
    return Template(text).substitute(scenario_id=scenario_id(context))
