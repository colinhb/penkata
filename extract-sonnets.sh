#!/bin/sh -e

# Get the directory where the script is located
script_dir=$(dirname "$0")
input_file="$script_dir/pg1041.txt"
output_dir="$script_dir/out"

# Check if input file exists
if [ ! -f "$input_file" ]; then
    echo "Error: Cannot find input file: $input_file" >&2
    exit 1
fi

# Create output directory if it doesn't exist
mkdir -p "$output_dir"

# Process the file with awk for better pattern matching and handling
awk -v output_dir="$output_dir" '
    # Function to convert Roman numeral to decimal
    function roman_to_decimal(roman) {
        # Define Roman numeral values
        values["I"] = 1
        values["V"] = 5
        values["X"] = 10
        values["L"] = 50
        values["C"] = 100

        roman = toupper(roman)
        decimal = 0
        prev_value = 0

        # Process right to left
        for (i = length(roman); i > 0; i--) {
            curr_value = values[substr(roman, i, 1)]
            if (curr_value >= prev_value) {
                decimal += curr_value
            } else {
                decimal -= curr_value
            }
            prev_value = curr_value
        }
        return decimal
    }

    # If line is empty and we have a sonnet, write it
    /^[[:space:]]*$/ {
        if (length(sonnet) > 0 && length(number) > 0) {
            decimal = roman_to_decimal(number)
            outfile = sprintf("%s/%d-%s.txt", output_dir, decimal, number)
            printf "%s", sonnet > outfile
            close(outfile)
            sonnet = ""
            number = ""
        }
        next
    }

    # If line contains only Roman numerals (and possibly whitespace)
    /^[[:space:]]*[IVXLC]+[[:space:]]*$/ {
        # Store the trimmed number
        number = $0
        sub(/^[[:space:]]+/, "", number)
        sub(/[[:space:]]+$/, "", number)
        next
    }

    # If we have a current sonnet number, collect lines
    length(number) > 0 {
        if (length(sonnet) == 0) {
            sonnet = $0
        } else {
            sonnet = sonnet "\n" $0
        }
    }

    # At end of file, write final sonnet if any
    END {
        if (length(sonnet) > 0 && length(number) > 0) {
            decimal = roman_to_decimal(number)
            outfile = sprintf("%s/%d-%s.txt", output_dir, decimal, number)
            printf "%s", sonnet > outfile
            close(outfile)
        }
    }
' "$input_file"

echo "Sonnets have been extracted to: $output_dir"
