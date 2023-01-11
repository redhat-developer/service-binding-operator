import os
from string import Template


def scenario_id(context):
    return f"{os.path.basename(os.path.splitext(context.scenario.filename)[0]).lower()}-{context.scenario.line}"


def substitute_scenario_id(context, text="$scenario_id"):
    return Template(text).substitute(scenario_id=scenario_id(context))


def get_env(env):
    value = os.getenv(env)
    assert env is not None, f"{env} environment variable needs to be set"
    print(f"{env} = {value}")
    return value
