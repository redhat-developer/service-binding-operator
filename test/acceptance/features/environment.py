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

import subprocess
import os


def before_all(_context):
    start_sbo = os.getenv("TEST_ACCEPTANCE_START_SBO")
    assert start_sbo is not None, "TEST_ACCEPTANCE_START_SBO is not set. It should be one of local, remote or operator-hub"
    assert start_sbo in {"local", "remote", "operator-hub"}, "TEST_ACCEPTANCE_START_SBO should be one of local, remote or operator-hub"

    if start_sbo == "local":
        assert not os.getenv("TEST_ACCEPTANCE_SBO_STARTED").startswith("FAILED"), "TEST_ACCEPTANCE_SBO_STARTED shoud not be FAILED."
    elif start_sbo == "remote":
        sbo_namespace = os.getenv("SBO_NAMESPACE")
        assert sbo_namespace is not None, "SBO_NAMESPACE is required but not set."
        _context.sbo_namespace = sbo_namespace
    else:
        assert False, f"TEST_ACCEPTANCE_START_SBO={start_sbo} is currently unsupported."


def before_scenario(_context, _scenario):
    code, output = subprocess.getstatusoutput('kubectl get project default -o jsonpath="{.metadata.name}"')
    assert code == 0, f"Checking connection to OS cluster by getting the 'default' project failed: {output}"
