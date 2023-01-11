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
from behave.model_core import Status

import os
import semver

cmd = Command()

if os.getenv("DELETE_NAMESPACE") in ["always", "never", "keepwhenfailed"]:
    delete_namespace = os.getenv("DELETE_NAMESPACE")
else:
    delete_namespace = "keepwhenfailed"


def before_all(_context):
    if ctx.cli == "oc":
        oc_ver = ctx.cli_version
        assert semver.compare(oc_ver, "4.5.0") > 0, f"oc version is required 4.5+, but is {oc_ver}."

        namespace = os.getenv("TEST_NAMESPACE")
        output, code = cmd.run(f"oc project {namespace}")
        assert code == 0, f"Cannot set default namespace to {namespace}, reason: {output}"

    start_sbo = os.getenv("TEST_ACCEPTANCE_START_SBO")
    assert start_sbo is not None, "TEST_ACCEPTANCE_START_SBO is not set. It should be one of local, remote or operator-hub"
    assert start_sbo in {"local", "remote", "operator-hub", "scenarios"}, "TEST_ACCEPTANCE_START_SBO should be one of local, remote or operator-hub"

    if start_sbo == "local":
        assert not str(os.getenv("TEST_ACCEPTANCE_SBO_STARTED")).startswith("FAILED"), "TEST_ACCEPTANCE_SBO_STARTED shoud not be FAILED."
    elif start_sbo == "remote":
        output, code = cmd.run(
            f"{ctx.cli} get deployment --all-namespaces -o json"
            + " | jq -rc '.items[] | select(.metadata.name == \"service-binding-operator\").metadata.namespace'")
        assert code == 0, f"Non-zero return code while trying to detect namespace for SBO: {output}"
        output = str(output).strip()
        assert output != "", "Unable to find SBO's deployment in any namespace."
        _context.sbo_namespace = output
    elif start_sbo == "scenarios":
        print("INFO: The scenarios are responsible for installing and running SBO...")
    else:
        assert False, f"TEST_ACCEPTANCE_START_SBO={start_sbo} is currently unsupported."
    ctx.no_scenarios_failed = True


def before_scenario(_context, _scenario):
    _context.bindings = dict()
    output, code = cmd.run(f'{ctx.cli} get ns default -o jsonpath="{{.metadata.name}}"')
    assert code == 0, f"Checking connection to OS cluster by getting the 'default' project failed: {output}"


def after_scenario(_context, scenario):
    if "namespace" in _context:
        namespace = _context.namespace.name
    elif os.getenv("TEST_NAMESPACE"):
        namespace = os.getenv("TEST_NAMESPACE")
    else:
        print("No namespace set in context nor TEST_NAMESPACE env variable is set, skipping deletion")
        return
    if delete_namespace == "always":
        delete_current_namespace(namespace)
    elif (delete_namespace == "keepwhenfailed"):
        if scenario.status == Status.passed and (ctx.no_scenarios_failed):
            delete_current_namespace(namespace)
        else:
            ctx.no_scenarios_failed = False
            print(f"Deleting namespace {namespace} skipped, since one of the scenarios failed.")
    elif delete_namespace == "never":
        print(f"Namespace {namespace} deletion skipped.")


def delete_current_namespace(name):
    output, code = cmd.run(f"{ctx.cli} delete namespace {name} --ignore-not-found --timeout=1800s")
    assert code == 0, f"Deletion of namespace failed: {output}"
    print(f"Namespace {name} deleted.")
