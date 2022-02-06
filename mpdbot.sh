#!/bin/bash
echo $(mpc -f '/music'/%file% current)> nowplayingpath
echo $(mpc current)> nowplaying
scp nowplaying nowplayingpath USER@HOST:PATH
