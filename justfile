install:
  go install && cd daemonizer && go run daemonizer.go && cd ..

refresh-logs:
  truncate -s 0 /var/log/shades/*.log && tail -f /var/log/shades/shades.*log
  
tail-logs:
  tail -f /var/log/shades/shades.*log

