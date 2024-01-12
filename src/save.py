import sys
import os
import shutil

def log(message):
    print(message)

def saveFiles(SteamLibrary, LocalLibrary):
    if not os.path.exists("../SavePathDataset.txt"):
        quit("Data path does not exist")
    SavePaths = open("../SavePathDataset.txt", 'r').read().splitlines()
    numberOfGamesToSave = len(SavePaths)
    print(f"The save data of {numberOfGamesToSave} games.")
    for path in SavePaths:
        gameName, savePath = path.split("|")

        log(f"saveing the data for {gameName}")
        log(f"Still need to save {numberOfGamesToSave - len(SavePaths[SavePaths.index(path)])} games data")

        shutil.copytree(f"{SteamLibrary}{savePath}", f"{LocalLibrary}{savePath}")

if __name__ == "__main__":
    # example SteamLibrary path /media/<user>/<drive>/SteamLibrary/
    SteamLibrary = None
    print(len(sys.argv))
    if len(sys.argv) > 1:
        SteamLibrary = sys.argv[1]
    if SteamLibrary == None:
        quit("Please provide SteamLibrary path")

    # append stuff to the SteamLibrary path to get to the games
    SteamLibrary = f"{SteamLibrary}steamapps/common/"

    LocalLibrary = None
    if len(sys.argv) > 2:
        LocalLibrary = sys.argv[2]
    if LocalLibrary == None:
        print("Local library is not set default is being used")

    if LocalLibrary == None:
        LocalLibrary = "../SteamSaveLocal/"

    if not os.path.exists(LocalLibrary):
        os.mkdir(LocalLibrary)

    saveFiles(SteamLibrary, LocalLibrary)
