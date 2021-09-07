import os


def scenario_id(context):
    return f"{os.path.basename(os.path.splitext(context.scenario.filename)[0]).lower()}-{context.scenario.line}"
