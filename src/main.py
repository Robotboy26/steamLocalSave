import sys
import os
import shutil
import time
import argparse
import zipfile
import platform
from concurrent.futures import ThreadPoolExecutor, as_completed
import pdb

debugMode = False

def log(message):
    if debugMode:
        print(message)

def timeFormat(option):
    currentTime = time.time()
    formattedTime = time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(currentTime))
    if option.lower() == "save":
        formattedTime = f"{formattedTime}-auto" # Used to tell the backup removing function to only select auto generated backups for removal
    elif option.lower() == "restore":
        formattedTime = f"{formattedTime}-backup" # Used to tell the backup removing function to only select auto generated backups for removal
    return formattedTime

def generatePaths(steamLibrary, localLibrary, gameName, savePath, option):
    if not "~" in savePath:
        src = os.path.normpath(f"{steamLibrary}/{savePath}")
    else:
        src = os.path.normpath(f"{savePath}")
        src = os.path.expanduser(src)

    src = src.strip() # Used to remove bad whitespaces
    targ = os.path.normpath(f"{localLibrary}{gameName}/{timeFormat(option)}/{gameName}")
    return src, targ

# TODO add support for dry runs
def performCopy(src, targ, gameName, dryRun=False):
    if dryRun == False:
        print(f"Saving game files for '{gameName}'")
        shutil.copytree(src, targ)
        zipPath = targ
        log(f"ZipPath: {zipPath}")
        shutil.make_archive(zipPath, 'zip', targ)
        log(f"Creating zip archive of '{gameName}'")
        shutil.rmtree(targ)

def unzipFile(zipPath, extractTo):
    if not zipfile.is_zipfile(zipPath):
        quit("The provided file is not a valid ZIP file.")

    # Create the directory to extract files if it doesn't exist
    os.makedirs(extractTo, exist_ok=True)

    with zipfile.ZipFile(zipPath, 'r') as zipRef:
        zipRef.extractall(extractTo)

    log(f"Extracted: {zipPath} to {extractTo}")

def saveGame(steamLibrary, localLibrary, maxBackups, option, path):
    try:
        # log(path)
        gameName, savePath = path.split("|")
    except ValueError:
        log("Value Error")
        return ValueError

    if option == None:
        quit("Please select save or restore (-s or -r)")

    src, targ = generatePaths(steamLibrary, localLibrary, gameName, savePath, option) # targ will end in '-auto' if it is a auto generated save file or '-backup' if it is created during the restore process
    if not os.path.exists(f"{src}"):
        log(f"You do not appear to have the game '{gameName}'")
        return

    log(f"option: {option}")
    if option.lower() == "save":
        performCopy(src, targ, gameName)
        zipFiles = [f for f in os.listdir(f"{localLibrary}{gameName}") if f.endswith("auto")]
        zipFiles = sorted(zipFiles, reverse=True)
        log(f"You have {len(zipFiles)} backups for game: '{gameName}'")

        while len(zipFiles) > maxBackups: # While loop because if you lower the amount of backups that you want saved you want all the old ones deleted
            log(f"zipFiles {zipFiles}. Bool: {len(zipFiles) > maxBackups}")
            log(f"More than {maxBackups} backups for game '{gameName}'. Removing the oldest: {zipFiles[-1]}")
            oldestBackup = os.path.join(f"{localLibrary}{gameName}", zipFiles[-1])
            shutil.rmtree(oldestBackup)
            zipFiles.pop() # If you deleted to oldest aready you have to remove the oldest from the list

        log(f"Saved data for '{gameName}'")
        return True
    if option.lower() == "restore":
        performCopy(src, targ, gameName) # Create the backup save to make sure no data is overwritten
        zipFiles = [f for f in os.listdir(f"{localLibrary}{gameName}") if f.endswith("auto")] # Only restore from autosaves
        zipFiles = sorted(zipFiles, reverse=True)
        log(f"Selecting the latest backup out of {len(zipFiles)} for game: '{gameName}'")
        latestBackup = os.path.join(f"{localLibrary}{gameName}", zipFiles[0])
        log(f"Restoring backup '{latestBackup}' to game files")
        zipPath = f"{latestBackup}/{gameName}.zip"
        shutil.rmtree(src)
        unzipFile(zipPath, src) # Unzip latest backup into the game save data location.
    # TODO add an option to remove the game data and related directorys after saving to clear up space on a computer.
        
    return None

def saveGames(steamLibrary, localLibrary, maxBackups, option):
    try:
        if not debugMode:
            readlines = open(f"../SavePathDataset-{platform.system()}.txt", 'r').read().splitlines()
        else:
            readlines = open(f"../SavePathDataset-{platform.system()}-Debug.txt", 'r').read().splitlines()
    except:
        quit(f"You do not have any datasets for platform: {platform.system()}")
    savePaths = []
    for path in readlines:
        if "#" in path and not path.startswith("#"):
            path = path.split("#") # This is for end of the line comments
            path = path[0]
        if not path.startswith("#") and not path == "": # If not a comment and not empty
            savePaths.append(path)
    numberOfGamesToSave = len(savePaths)
    log(f"Searching {numberOfGamesToSave} game save data locations.")
    if debugMode == True:
        for path in savePaths:
            saveGame(steamLibrary, localLibrary, maxBackups, option, path)
    else:
        with ThreadPoolExecutor() as executor:
            futures = {
                    executor.submit(saveGame, steamLibrary, localLibrary, maxBackups, option, path): path for path in savePaths
                    }

        for future in as_completed(futures):
            path = futures[future]
            try:
                result = future.result()  # This retrieves the result of the call
                if result:
                    print(f"Successfully saved game for path: {path}.")
            except Exception as e:
                print(f"Error occurred while saving game for path: {path}. Exception: {e}")

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
    # Example steamLibrary path /media/<user>/<drive>/steamLibrary/
    # Add argParser with required steamLibrary path but other optional
    steamLibrary = None


    if args.debug:
        debugMode = True
        log(args)

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

    log(args)

    saveGames(args.steamLibraryPath, args.localLibrary, args.backups, option)
