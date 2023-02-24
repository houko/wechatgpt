#!/bin/bash
basepath=$(cd `dirname $0`; pwd)

nohup ./wechatbot >> $basepath/run.log &
