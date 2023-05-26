#!/bin/bash

# Create a new tmux session
tmux new-session -d -s mysession

# Split the terminal into 2x3 layout
tmux split-window -h
tmux split-window -h
tmux select-layout even-horizontal
tmux split-window -v
tmux select-pane -t 0
tmux split-window -h
tmux split-window -h

# Define the commands
commands=(
  "./main -layerLogo=APP -deviceId=0"
  "./main -layerLogo=LNK -deviceId=0"
  "./main -layerLogo=PHY -deviceId=0"
  "./main -layerLogo=APP -deviceId=1"
  "./main -layerLogo=LNK -deviceId=1"
  "./main -layerLogo=PHY -deviceId=1"
)

# Execute the commands in each pane
pane_index=0
for command in "${commands[@]}"; do
  tmux send-keys -t $pane_index "$command" Enter
  pane_index=$((pane_index + 1))
done

# Attach to the tmux session
tmux attach-session -t mysession
