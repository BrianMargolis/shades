init-logs:
  # creating a log directory if it doesn't exist already
  if [ ! -d /var/log/shades ]; then sudo mkdir /var/log/shades && sudo chown $(whoami) /var/log/shades; fi

install: init-logs
  # install the binary and run the daemonizer utility
  go install && cd daemonizer && go run daemonizer.go && cd ..

refresh-logs:
  # clear the log files and start tailing them
  truncate -s 0 /var/log/shades/*.log && tail -f /var/log/shades/shades.*log
  
tail-logs:
  # tail the log files
  tail -f /var/log/shades/shades.*log

