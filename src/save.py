import sys
import os
import shutil
import time
import argparse
import zipfile
import platform
import concurrent.futures
import pdb

debugMode = False

def log(message):
    if debugMode:
        print(message)

def timeFormat():
    currentTime = time.time()
    formattedTime = time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(currentTime))
    return formattedTime

def generatePaths(steamLibrary, localLibrary, gameName, savePath, backupPath):
    if not "~" in savePath:
        src = os.path.normpath(f"{steamLibrary}/{savePath}")
    else:
        src = os.path.normpath(f"{savePath}")
        src = os.path.expanduser(src)

    targ = os.path.normpath(f"{localLibrary}{gameName}/{timeFormat()}/{backupPath}")
    return src, targ

def performCopy(src, targ, gameName, dryRun=False):
    if dryRun == False:
        log(f"Saving files for '{gameName}'")
        shutil.copytree(src, targ)
        zipPath = targ
        shutil.make_archive(zipPath, 'zip', targ)
        log(f"Creating zip archive of '{gameName}'")
        shutil.rmtree(targ)

def saveGame(steamLibrary, localLibrary, maxBackups, option, path):
    try:
        log(path)
        gameName, savePath, backupPath = path.split("|")
    except ValueError:
        log("Value Error")
        quit("Value Error")

    if option == None:
        quit("Please select save or restore")

    src, targ = generatePaths(steamLibrary, localLibrary, gameName, savePath, backupPath)
    if not os.path.exists(f"{src}"):
        log(f"You do not appear to have the game '{gameName}'")
        return

    if not os.path.exists(f"{localLibrary}{gameName}"): # If save path does not exist
        # This is here because you get things like <folder>/../../<folder> and this errors for some reason
        performCopy(src, targ, gameName)
        log(f"Saved data for {gameName}")
    else:
        if option.lower() == "save":
            zipFiles = [f for f in os.listdir(f"{localLibrary}{gameName}") if f.endswith(".zip")]
            zipFiles = sorted(zipFiles)
            log(f"You have {len(zipFiles)} backups for game: '{gameName}'")
        
            while len(zipFiles) > maxBackups: # While loop because if you lower the amount of backups that you want saved you want all the old ones deleted
                log(f"More than {maxBackups} backups for game '{gameName}'. Removing the oldest: {zipFiles[-1]}")
                oldestBackup = os.path.join(f"{localLibrary}{gameName}", zipFiles[-1])
                os.remove(oldestBackup)
            performCopy(src, targ, gameName)
            log(f"Saved data for {gameName}")
        if option.lower() == "restore":
            quit("Not yet implimented") # TODO

    return

def saveGames(steamLibrary, localLibrary, maxBackups, option):
    log(platform.system())
    try:
        readlines = open(f"../SavePathDataset-{platform.system()}.txt", 'r').read().splitlines()
    except:
        quit(f"You do not have any datasets for platform: {platform.system()}")
    savePaths = []
    for path in readlines:
        if "**" in path and not path.startswith("**"):
            path = path.split("**") # This is for end of the line comments
            path = path[0]
        if not path.startswith("**") and not path == "": # If not a comment and not empty
            savePaths.append(path)
    numberOfGamesToSave = len(savePaths)
    log(f"Searching {numberOfGamesToSave} game save data locations.")
    if not debugMode:
        with concurrent.futures.ThreadPoolExecutor() as executor:
            futures = [executor.submit(saveGame, steamLibrary, localLibrary, maxBackups, option, path) for path in savePaths]
    else:
        for path in savePaths:
            saveGame(steamLibrary, localLibrary, maxBackups, option, path)

def pushFiles():
    pass

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Used to save and restore steam game save data')
    parser.add_argument('steamLibraryPath', nargs='?', help='The path to your steam library')
    parser.add_argument('-s', '--save', action='store_true', help='Save Steam game data')
    parser.add_argument('-r', '--restore', action='store_true', help='Restore Steam game data')
    parser.add_argument('-l', '--localLibrary', nargs='?', help='Location to store or restore Steam data from or to')
    parser.add_argument('-b', '--backups', type=int, help='The number of Steam game backups to store during saving')
    parser.add_argument('-d', '--debug', action='store_true', help='Used by developers to debug the program')
    args = parser.parse_args()
    # example steamLibrary path /media/<user>/<drive>/steamLibrary/
    # add argParser with required steamLibrary path but other optional
    steamLibrary = None
    print(args)

    if args.debug:
        debugMode = True

    if args.backups == None:
        args.backups = 2 # default
    if args.localLibrary == None:
        args.localLibrary = "../SteamSaveLocal/"

    if not os.path.exists(args.localLibrary):
        os.mkdir(args.localLibrary)

    option = None
    if args.save:
        option = "save"

    if args.restore:
        option = "restore"

    saveGames(args.steamLibraryPath, args.localLibrary, args.backups, option)
