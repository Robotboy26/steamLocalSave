import sys
import os
import shutil
import time
import argparse
import zipfile
import platform
import concurrent.futures
import pdb

def log(message):
    print(message)

def timeFormat():
    currentTime = time.time()
    formattedTime = time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(currentTime))
    return formattedTime

def generatePaths(SteamLibrary, gameName, savePath, backupPath):
    if not "~" in savePath:
        src = os.path.normpath(f"{SteamLibrary}/{savePath}")
    else:
        src = os.path.normpath(f"{savePath}")
        src = os.path.expanduser(src)

    targ = os.path.normpath(f"{LocalLibrary}{gameName}/{timeFormat()}/{backupPath}")
    return src, targ

def performCopy(src, targ, gameName, dryRun=False):
    if dryRun == False:
        log(f"Saving files for '{gameName}'")
        shutil.copytree(src, targ)
        zipPath = targ
        shutil.make_archive(zipPath, 'zip', targ)
        log(f"Creating zip archive of '{gameName}'")
        shutil.rmtree(targ)

def saveGame(SteamLibrary, LocalLibrary, maxBackups, option, path):
    try:
        gameName, savePath, backupPath = path.split("|")
    except ValueError:
        return

    # log(f"Still need to save {numberOfGamesToSave - SavePaths.index(path) - 1} games data")

    src, targ = generatePaths(SteamLibrary, gameName, savePath, backupPath)
    if not os.path.exists(f"{src}"):
        # log(f"You do not appear to have the game '{gameName}'")
        return

    if not os.path.exists(f"{LocalLibrary}{gameName}"): # is save path does not exist
        # this is here because you get things like <folder>/../../<folder> and this errors
        performCopy(src, targ, gameName)
        log(f"Saved data for {gameName}")
    else:
        if option.lower() == "all":
            zipFiles = [f for f in os.listdir(f"{LocalLibrary}{gameName}") if f.endswith(".zip")]
            zipFiles = sorted(zipFiles)
            log(f"You have {len(zipFiles)} backups for game: '{gameName}'")
        
            while len(zipFiles) > maxBackups:
                log(f"More than {maxBackups} backups for game '{gameName}'. Removing the oldest: {zipFiles[-1]}")
                oldestBackup = os.path.join(f"{LocalLibrary}{gameName}", zipFiles[-1])
                os.remove(oldestBackup)
            performCopy(src, targ, gameName)
            log(f"Saved data for {gameName}")

    return

def saveGames(SteamLibrary, LocalLibrary, maxBackups, option):
    if not os.path.exists("../SavePathDatasetLinux.txt"):
        quit("You do not have any datasets.")
    if platform.system() == "Linux":
        readlines = open("../SavePathDatasetLinux.txt", 'r').read().splitlines()
    else:
        quit("you only have the database for Linux.")
    SavePaths = []
    for path in readlines:
        if "**" in path and not path.startswith("**"):
            path = path.split("**") # this is for end of the line comments
            path = path[0]
        if not path.startswith("**") or not path == "":
            SavePaths.append(path)
    numberOfGamesToSave = len(SavePaths)
    log(f"The save data of {numberOfGamesToSave} games.")
    with concurrent.futures.ThreadPoolExecutor() as executor:
        futures = [executor.submit(saveGame, SteamLibrary, LocalLibrary, maxBackups, option, path) for path in SavePaths]

def pushFiles():
    pass

if __name__ == "__main__":
    # example SteamLibrary path /media/<user>/<drive>/SteamLibrary/
    # add argParser with required SteamLibrary path but other optional
    SteamLibrary = None
    maxBackups = 2 # default
    print(len(sys.argv))
    if len(sys.argv) > 2:
        option = sys.argv[1]
        SteamLibrary = sys.argv[2]
    if SteamLibrary == None:
        quit("Please provide SteamLibrary path")

    # append stuff to the SteamLibrary path to get to the games
    SteamLibrary = f"{SteamLibrary}steamapps/"

    LocalLibrary = None
    if len(sys.argv) > 3:
        LocalLibrary = sys.argv[3]
    if LocalLibrary == None:
        print("Local library is not set default is being used")

    if LocalLibrary == None:
        LocalLibrary = "../SteamSaveLocal/"

    if not os.path.exists(LocalLibrary):
        os.mkdir(LocalLibrary)

    saveGames(SteamLibrary, LocalLibrary, maxBackups, option)
