#!/bin/bash
while read p; do
	pref=$(echo $p | cut -d '/' -f 6)	
	./extract -e $p -C ../data/coocsW10/$pref.cooc -U ../data/unigrams/merged.unigram -v 50000 -w 10 -option cooc
	gzip ../data/coocsW10/$pref.cooc
	exit 0
done<$1

