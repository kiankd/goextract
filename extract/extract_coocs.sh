#!/bin/bash
if [[ -z $1 && -z $2 && -z $3 && -z $4 ]]; then
	echo "Need to pass path to unigram, path to .gz files, experiment name, and extra args string (including window option!)"
	exit 0
fi

# Extract cooccurrence data from the files in the input directory.
# For a 275MB .txt.gz file with context window w=5, this runs in about 4 minutes.
# In $4 please pass any additional arguments you would like ./extract to take in.
# The most important is, of course, window size/path to window!

exp=$3-run
i=0
for f in $2/*.gz; do
	echo "RUNNING: ./extract -option cooc -e $f -U $1 -C ../data/coocs/$exp/$i.cooc $4"
	./extract -option cooc -e $f -U $1 -C ../data/coocs/$exp/$i.cooc $4
	i=$((i+1))
done


