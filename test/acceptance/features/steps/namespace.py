from environment import ctx

from command import Command


class Namespace(object):
    def __init__(self, name):
        self.name = name
        self.cmd = Command()

    def create(self):
        output, exit_code = self.cmd.run(f"{ctx.cli} create namespace {self.name}")
        assert exit_code == 0, f"Unexpected output when creating namespace: '{output}'"
        return True

    def is_present(self):
        _, exit_code = self.cmd.run(f'{ctx.cli} get ns {self.name}')
        return exit_code == 0
