# onegw


找到进程 ps -ef|grep onegw

pgrep onegw | xargs kill -9


ps -efww | grep -w 'onegw'


#!/bin/bash

go run /root/onegw/cmd/onegw/main.go