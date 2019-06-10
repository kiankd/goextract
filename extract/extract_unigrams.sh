#!/bin/bash
if [[ -z $1 && -z $2 ]]; then
	echo "Must pass with argument for name of this run (\$1)!"
	echo "Must also pass with target directory! (\$2)!"
	exit 0
fi

# Use this script to extract the unigram from the data, and then merge it.
# If you want, use $3 as the argument for the vocabulary size!
# It takes about 25 seconds to extract a unigram from a 275MB .txt.gz file.
# It assumes that you are storing your unigrams in ../data/unigrams.

if [[ -z $3 ]]; then
	vocabsize=50000
else
	vocabsize=$3
fi

mkdir ../data/unigrams/$1-run
i=0
for f in $2/*.gz; do
	./extract -option unigram -e $f -U ../data/unigrams/$1-run/$i.unigram; 
	i=$((i+1))
done
./extract -option unigram-merge -U ../data/unigrams/$1-run/ -v $vocabsize
