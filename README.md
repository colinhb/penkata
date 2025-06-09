# Penkata

The goal of this project is to find short passages of text from a corpus which have highest unique number of the most common bigrams in that corpus.

In other words, we want to take a corpus of text, and find the frequency of all bigrams in that corpus, and then for a given passage length, rank all passages in the corpus based on a weighted score of the number of unqiue bigrams contained in the passage, where the weights are the frequency of each bigram in the entire corpus.

Notably, we are treating whitespace transitions as bigrams. The string `Foo Bar Baz` contains the bigrams `Fo oo o_ _B Ba ar r_ _B Ba az` and the bigram counts:

```
Fo	1
oo	1
o_	1
_B	2
Ba	2
ar	1
r_	1
az	1
```

The rough motivation for this project is to find the ideal text to practice penmanship--thus Penkata: pen + kata.

## Gutenberg Results

Using bigram counts from [Project Gutenberg](https://www.gutenberg.org), here are the two top-scoring passages of 300 characters with normalized (occurances/total) bigram weighting:

> The weight of the attack fell upon the flank, which was completely crushed, and a part of the intrenchments
> **with their guns fell into the hands of the National forces. 
> In the night Hood retreated a mile or two, to another line on the hills, made some new dispositions, and awaited attack. 
> He was seriously embarrassed by the absence of a large part of Forrest's cavalry, which should have been protecting his** 
> flanks.
> 
> Johnson, Rossiter. (1894). Campfire and Battlefield: An Illustrated History of the Campaigns and Conflicts of the Great Civil War.

> It would suffice, however, if I sighted
> **it from a distance; that would give me a sufficiently exact position. 
> So I hardened my heart and laid my course for the western extremity of the bank.
> By now the storm was howling like hell let loose, the squalls slung volleys of hailstones down on us. 
> The seas raged against our littleship, which** 
> still shouldered them from her gallantly.
> 
> Spindler, Karl. (1921). Gun Running for Casement in the Easter Rebellion, 1916. (Translated from the German by Montgomery, W., and McGrath, E.H.)

The sections **in bold** are the highest-scoring passages. I've included the remainder of the cut-off sentences for readability.

Bigram counts from Project Gutenberg are available at `out/bigrams/gutenberg.tsv` and passages scored with various parameters (character limits, weighting) at `out/passages/gutenberg.tsv`

## Project Structure

The project follows standard Go project layout with the following organization:

- `cmd/`: Contains the main executable programs
  - `bigrams/`: Command to extract and count bigrams from text files
  - `passages/`: Command to find and rank passages by bigram scores
- `pkg/`: Reusable packages
  - `myfuncs/`: Generic utility functions (Map, Filter, Reduce, etc.)
  - `mytypes/`: Generic type definitions (Set, etc.)
  - `penkata/`: Core functionality for text analysis
- `sonnets/`: Sample text corpus (Shakespeare's sonnets)
- `out/`: Output directory for generated files
  - `bigrams/`: Generated bigram count files (.tsv)
  - `passages/`: Passages ranked by bigram scores (.tsv)
- `bin/`: Build output directory for executable binaries

## Building

This project uses a Plan 9 style mkfile for convenience. To build all executables, assuming you are using [plan9port](https://9fans.github.io/plan9port/):

```sh
9 mk all
```

This will create the binaries in the `bin/` directory. 

## Usage

### Generating Bigram Counts

First, generate bigram counts from a corpus:

```sh
./bin/bigrams -d ./sonnets > ./out/bigrams/sonnets.tsv
```

This produces a TSV file with each bigram and its frequency in the corpus.

### Finding Best Passages

To find passages with the highest bigram scores:

```sh
./bin/passages -f ./out/bigrams/sonnets.tsv -d ./sonnets -c 200 -n 5 -w raw
```

Parameters:
- `-f`: Path to TSV file with bigram counts (default: "sonnets-bigrams.tsv")
- `-d`: Directory containing text files to analyze (required)
- `-c`: Maximum character length for passages (default: 200, can specify multiple: `-c 150 -c 300`)
- `-n`: Number of top passages to display (default: 50)
- `-o`: Output file for results in TSV format (optional, default: stdout)
- `-v`: Enable verbose output with statistics during processing
- `-w`: Weight transformation type (raw, log1p, normal) (can specify multiple: `-w log1p -w normal`)

### How It Works

The passage finder:
1. Loads bigram weights from the specified TSV file
2. Walks through the directory, processing each text file
3. For each file, maintains sliding windows of text
4. Scores windows based on the corups frequency of unique bigrams, possibly normalized or otherwise transformed
5. Ranks and returns the highest-scoring passages

### Example Output

When writing to standard output (default), the results are formatted with section headers:

```
=== Results for 150 character passages (raw) ===
transform       maxChar path    score   size    text
raw     150     sonnets/110-CX.txt      127221.00       150     heart another youth, And worse essays prov’d thee my best of love. Now all is done, save what shall have no end: Mine appetite I never more will grind

=== Results for 200 character passages (raw) ===
transform       maxChar path    score   size    text
raw     200     sonnets/12-XII.txt      139820.00       200     of leaves, Which erst from heat did canopy the herd, And summer’s green all girded up in sheaves, Borne on the bier with white and bristly beard, Then of thy beauty do I question make, That thou among
```

When using the `-o` flag to output to a file, results are provided in tab-separated format without section headers:

```
transform	maxChar	path	score	size	text
raw	150	sonnets/110-CX.txt	127221.00	150	heart another youth, And worse essays prov’d thee my best of love. Now all is done, save what shall have no end: Mine appetite I never more will grind
raw	200	sonnets/12-XII.txt	139820.00	200	of leaves, Which erst from heat did canopy the herd, And summer’s green all girded up in sheaves, Borne on the bier with white and bristly beard, Then of thy beauty do I question make, That thou among
```

## Requirements

- Go 1.24 or later
- Plan 9 port (for Plan 9 style mkfile)
