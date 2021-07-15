"""
before_step(context, step), after_step(context, step)
    These run before and after every step.
    The step passed in is an instance of Step.
before_scenario(context, scenario), after_scenario(context, scenario)
    These run before and after each scenario is run.
    The scenario passed in is an instance of Scenario.
before_feature(context, feature), after_feature(context, feature)
    These run before and after each feature file is exercised.
    The feature passed in is an instance of Feature.
before_all(context), after_all(context)
    These run before and after the whole shooting match.
"""

from steps.command import Command
from steps.environment import ctx

import os

cmd = Command()


def before_all(_context):
    if ctx.cli == "oc":
        output, code = cmd.run("oc version | grep Client")
        assert code == 0, f"Checking oc version failed: {output}"

        oc_ver = output.split()[2]
        assert oc_ver >= "4.5", f"oc version is required 4.5+, but is {oc_ver}."

        namespace = os.getenv("TEST_NAMESPACE")
        output, code = cmd.run(f"oc project {namespace}")
        assert code == 0, f"Cannot set default namespace to {namespace}, reason: {output}"

    output, code = cmd.run(
        f"{ctx.cli} get deployment --all-namespaces -o json"
        + " | jq -rc '.items[] | select(.metadata.name == \"service-binding-operator\").metadata.namespace'")
    assert code == 0, f"Non-zero return code while trying to detect namespace for SBO: {output}"
    output = str(output).strip()
    assert output != "", "Unable to find SBO's deployment in any namespace."
    _context.sbo_namespace = output


def before_scenario(_context, _scenario):
    _context.bindings = dict()
    output, code = cmd.run(f'{ctx.cli} get ns default -o jsonpath="{{.metadata.name}}"')
    assert code == 0, f"Checking connection to OS cluster by getting the 'default' project failed: {output}"
