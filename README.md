# goextract
## Go code for cooccurrence extraction of corpus statistics.

This software is used to extract the cooccurrence statistics within a corpus of text. You will need:
- go (https://golang.org/doc/install), version 1.10 or higher
- to have a basic understanding of bash scripting

### Preliminary set up.
We can think of statistics extraction in terms of 5 separate steps: 
*Preparation; Unigram extraction; Unigram merging; Cooccurrence extraction; Cooccurence merging.*

- Step 1. I have a big .txt file that is defined with two notions of separation: spaces indicate new tokens, and newlines indicate new documents. Maybe this file is 30 GB.
- Step 1.1. I want to divide that file up into smaller 1 GB pieces (this will later facilitate generalized multiprocessing for extremely rapid extraction); use the bash command $split to do this!
- Step 1.2. I want to store things efficiently, so gzip the divided files -- the Go code assumes its given gzipped files anyway.

### Unigram extraction.
- Step 2. Now, I want to know what the unigram statistics of this corpus are -- this will produce an encoder-decoder structure that is required for cooccurrence extraction. Suppose the divided data is in a directory `divided/`, and we want to store results files into the directory `unigrams/`, then we do:

```bash
i=0; 
for f in divided/*.gz; do
	i=$((i+1)); 
	./extract -option unigram -e $f -U unigrams/$i.unigram; 
done
```

- Step 3. This script will produce a bunch of sub-unigram files for each file in `divided/`. But, we would rather have a single merged unigram file; additionally we need to specify what the desired vocabulary size should be -- 50,000 is often a good number! This is easy:

`./extract -option unigram-merge -U unigrams/ -v 50000`

### Cooccurrence extraction.
- Step 4. Now we have the good boy file `unigrams/merged.unigram` which stores our vocabulary, and the unigram statistics. This will be used to help us extract cooccurrences! However, when doing cooccurrence extraction there is one fundamental consideration: _how do we define the context_?
- Step 4a. We could use *dynamic context window weighting*, like Word2vec; in this case, we just use the argument `-w W`, where W is the desired context window size (typically in the range of 2-10); note, the larger W is, the longer the extraction will take!
- Step 4b. We could use a *generalized context window file*; e.g., perhaps we want to define an assymetric context window with our own desired weights. This is done by passing `-window /path/to/window_file.w`; examples of how the .w file should be written are found in `data/test_data/`, which includes left and right assymetric examples.

- Step 4.1. Do the extraction! Let's suppose you are using a basic 5-token left-right context window, and we are storing temporary `.cooc` files into a directory called `coocs/`:

```bash
i=0; 
for f in divided/*.gz; do
    i=$((i+1));  
    ./extract -option cooc -e $f -U unigrams/merged.unigram -C coocs/$i.cooc -w 5;   
done
```

- Step 5. Merge the results from extraction!

`./extract -option cooc-merge -C coocs/`

### Final comments.
- We now have a file called `coocs/merged.cooc`. This file stores all of the cooccurrence information in the corpus according to the definition of window size, and exists only with respect to the vocabulary encoding defined by the unigram file used during extraction (`unigrams/merged.unigram`). It is structured as, for each line: *term_i context_j Nij*, where i and j are the codes defined in the unigram file that map to the unigram file's string.
- *Concurrency pattern*: instead of using a for loop to make each .cooc file one at a time, we could multiprocess this and divide responsibility to just iterate over K .gz files, rather than all N. By doing so you can considerably speed up running time; e.g., dividing into 4 simultaneous processes will reduce runtime by x4.
- *Full path pattern*: at the current state of this project, everything requires the full path in order to run properly; so, always use the full path to any directory or file when using it; e.g., instead of doing `-C coocs/` you will probably need to do `-C /home/rldata/hilbert-data/coocs`, etc.
- *RAM usage*: this code will use a considerable amount of RAM during _Step 4.1_, and it is highly concurrent within `./extract`; therefore, be careful when using on a big server as it will use all available cores (but will be very fast). If you have more than 32 GB of RAM you should be pretty much good; if you have more than 64 GB of RAM then you will certainly be fine.
- *Smart usage*: step 4.1 is the only expensive operation, every other operation can be done in the space of a few seconds/minutes; therefore, when thinking about parallelizing, only consider it with respect to step 4.1 --- it is not necessary to parallelize the unigram extraction (although you could do so with exactly the same pattern as you would do for 4.1).




