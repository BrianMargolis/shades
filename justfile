install:
  # install the binary and register launchd agents
  go install && shades install

uninstall:
  # stop and remove launchd agents
  shades uninstall

refresh-logs:
  # clear the log files and start tailing them
  truncate -s 0 ~/.shades/logs/*.log && tail -f ~/.shades/logs/*.log

tail-logs:
  # tail the log files
  tail -f ~/.shades/logs/*.log
