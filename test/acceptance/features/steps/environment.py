import os
from steps.command import Command


class Environment(object):
    cli = "oc"
    cli_version = None

    def __init__(self, cli, cli_version=None):
        self.cli = cli
        self.cli_version = cli_version


# This is a global context (complementing behave's context)
# to be accesible from any place, even where behave's context is not available.
global ctx
cmd = Command()
cli = os.getenv("TEST_ACCEPTANCE_CLI", "oc")
cli_version = None
if cli == "oc":
    output, code = cmd.run("oc version --client | grep Client")
    assert code == 0, f"Checking {cli} version failed: {output}"
    cli_version = output.split()[2]
elif cli == "kubectl":
    output, code = cmd.run("kubectl version -o json | jq -rc '.clientVersion.gitVersion'")
    assert code == 0, f"Checking {cli} version failed: {output}"
    cli_version = output.split(sep="v")[1]

ctx = Environment(cli, cli_version)
