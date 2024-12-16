# TODO turning this into a github action that will sort it on each commit so that it does not have to be done manually?
import sys

def sort_file(filename):
    try:
        with open(filename, 'r') as file:
            lines = file.readlines()
        
        sortedLines = sorted(line.strip() for line in lines)
        
        for line in sortedLines: # Copy and paste this into a new file
            print(line)

    except FileNotFoundError:
        print(f"Error: The file '{filename}' was not found.")
    except Exception as e:
        print(f"An error occurred: {e}")

if __name__ == '__main__':
    if len(sys.argv) != 2:
        print("Usage: python3 sort_file.py <filename>")
    else:
        sort_file(sys.argv[1])
