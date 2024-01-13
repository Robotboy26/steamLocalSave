import shutil
import pdb
# this will sort the save dataset
def main():
    shutil.copy("../SavePathDatasetLinux.txt", "../SavePathDatasetLinux.txt.backup")
    readlines = open("../SavePathDatasetLinux.txt", 'r').read().splitlines()
    SavePaths = []
    # ignore comments
    for path in readlines:
        if not path.startswith("//"):
            SavePaths.append(path)
    numberOfGamesToSave = len(SavePaths)
    print(f"Sorting {numberOfGamesToSave} games.")
    SavePaths = sorted(SavePaths, reverse=True)
    sortedIndex = 0
    for x in range(len(readlines)):
        if not readlines[x].startswith("//"):
            readlines[x] = SavePaths[sortedIndex]
            sortedIndex += 1

    with open("../SavePathDatasetLinux.txt", 'w') as F:
        F.write("\n".join(readlines))

if __name__ == "__main__":
    main()
