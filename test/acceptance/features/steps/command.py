import subprocess
import time
import os


class Command(object):
    path = ""
    env = {}

    def __init__(self, path=None):
        if path is None:
            self.path = os.getcwd()
        else:
            self.path = path
        self.setenv("KUBECONFIG", os.getenv("KUBECONFIG"))
        self.setenv("PATH", os.getenv("PATH"))

    def setenv(self, key, value):
        self.env[key] = value

    def run(self, cmd, stdin=None):
        output = None
        exit_code = 0
        try:
            if stdin is None:
                output = subprocess.check_output(cmd, shell=True, stderr=subprocess.STDOUT, cwd=self.path, env=self.env)
            else:
                output = subprocess.check_output(cmd, shell=True, stderr=subprocess.STDOUT, cwd=self.path, env=self.env, input=stdin.encode("utf-8"))
        except subprocess.CalledProcessError as err:
            output = err.output
            exit_code = err.returncode
            print('ERROR MESSGE:', output)
            print('ERROR CODE:', exit_code)
        return output.decode("utf-8"), exit_code

    def run_wait_for_status(self, cmd, status, interval=20, timeout=180):
        cmd_output = None
        exit_code = -1
        start = 0
        while ((start + interval) <= timeout):
            cmd_output, exit_code = self.run(cmd)
            if status in cmd_output:
                return True, cmd_output, exit_code
            time.sleep(interval)
            start += interval
        print("ERROR: Time out while waiting for status message.")
        return False, cmd_output, exit_code
