# Plan 9 style mkfile for the penkata project

MKSHELL=rc

BIN=bin
TARGS=$BIN/bigrams $BIN/passages
GOFLAGS=-v
PKGS=`{ls pkg/myfuncs/*.go pkg/mytypes/*.go pkg/penkata/*.go}
CORPUS=gutenberg
MIRROR=rsync://gutenberg.pglaf.org/gutenberg
RSYNC_OPTS='-ah --progress --del'

# Default target builds all binaries
all:V: $TARGS

# Ensure bin directory exists
$BIN:
	mkdir -p $BIN

# Ensure counts directory exists
counts:
	mkdir -p counts

# Build the bigrams command
$BIN/bigrams: $BIN $PKGS ./cmd/bigrams/main.go
	go build $GOFLAGS -o $BIN/bigrams ./cmd/bigrams

# Build the passages command
$BIN/passages: $BIN $PKGS ./cmd/passages/main.go
	go build $GOFLAGS -o $BIN/passages ./cmd/passages

# Generate sonnets bigram counts
counts/sonnets.tsv: $BIN/bigrams counts
	./$BIN/bigrams -d ./sonnets/out > ./counts/sonnets.tsv

# Corpus target - verifies symbolic link exists and is accessible
corpus:V:
	# Check if gutenberg symbolic link exists
	# Note: using the the unix test command, not plan9port
	if(! u test -L $CORPUS) {
		echo 'Error: '$CORPUS' symbolic link does not exist.' >[1=2]
		echo 'Please create a symbolic link to your external storage:' >[1=2]
		echo '	ln -s /path/to/external/storage' $CORPUS >[1=2]
		exit 1
	}
	# Check if the link target exists and is a directory
	if(! u test -d $CORPUS) {
		echo 'Error: '$CORPUS' exists but target is not mounted or accessible.' >[1=2]
		echo 'Please ensure your external storage is mounted properly.' >[1=2]
		exit 1
	}
	echo 'Corpus target is accessible.'

# Sync the Gutenberg corpus locally (depends on corpus target)
rsync-corpus:V: corpus
	echo 'Starting rsync from' $MIRROR 'to' $CORPUS'...'
	rsync $RSYNC_OPTS $MIRROR $CORPUS
	if(~ $status '') {
		echo 'rsync completed successfully.'
	}
	if not {
		echo 'rsync encountered an error.' >[1=2]
		exit 1
	}

# Clean target removes binaries
clean:V:
	rm -f $TARGS

# Clean and rebuild everything
nuke:V: clean all
