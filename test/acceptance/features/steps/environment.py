import os


class Environment(object):
    cli = "oc"

    def __init__(self, cli):
        self.cli = cli


# This is a global context (complementing behave's context)
# to be accesible from any place, even where behave's context is not available.
global ctx
ctx = Environment(os.getenv("TEST_ACCEPTANCE_CLI", "oc"))
