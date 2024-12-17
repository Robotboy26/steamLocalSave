import sys
import os
import shutil
import time
import argparse
import yaml
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

def generatePaths(steamLibrary, localLibrary, gameName, savePaths, option):
    srcList = []
    for path in savePaths:
        if not "~" in path:
            src = os.path.normpath(f"{steamLibrary}/{path}")
        else:
            src = os.path.normpath(f"{path}")
            src = os.path.expanduser(src)

        src = src.strip() # Used to remove bad whitespaces
        srcList.append(src)
    targ = os.path.normpath(f"{localLibrary}{gameName}/{timeFormat(option)}/{gameName}")
    return srcList, targ

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

def saveGame(steamLibrary, localLibrary, maxBackups, option, path, uuid):
    try:
        # log(path)
        pathSplit = path.split("|")
        gameName = pathSplit[0]
        savePaths = pathSplit[1:] # Supporting multiple path search because it changes depending on the Linux distro
        # TODO find a order to sort these or figure out which os is being used with more precision than just Linux
    except ValueError:
        log("Value Error")
        return ValueError

    if option == None:
        quit("Please select save or restore (-s or -r)")

    srcList, targ = generatePaths(steamLibrary, localLibrary, gameName, savePaths, option) # targ will end in '-auto' if it is a auto generated save file or '-backup' if it is created during the restore process
    # Find the correct src path
    atLeastOneSrcExists = False
    foundSources = []
    for src in srcList:
        # Check for the uuid
        if ";" in src and uuid != None:
            src = src.replace(";", str(uuid))
        if os.path.exists(src):
            foundSources.append(src)

        if len(foundSources) == 1:
            atLeastOneSrcExists = True
            foundSrc = foundSources[0]
        elif len(foundSources) > 1:
            quit("Found multiple game save data Location")

    if not atLeastOneSrcExists:
        log(f"You do not appear to have the game '{gameName}'")
        return None

    log(f"option: {option}")
    if option.lower() == "save":
        performCopy(foundSrc, targ, gameName)
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
        performCopy(foundSrc, targ, gameName) # Create the backup save to make sure no data is overwritten
        zipFiles = [f for f in os.listdir(f"{localLibrary}{gameName}") if f.endswith("auto")] # Only restore from autosaves
        zipFiles = sorted(zipFiles, reverse=True)
        log(f"Selecting the latest backup out of {len(zipFiles)} for game: '{gameName}'")
        latestBackup = os.path.join(f"{localLibrary}{gameName}", zipFiles[0])
        log(f"Restoring backup '{latestBackup}' to game files")
        zipPath = f"{latestBackup}/{gameName}.zip"
        shutil.rmtree(foundSrc)
        unzipFile(zipPath, foundSrc) # Unzip latest backup into the game save data location.
    # TODO add an option to remove the game data and related directorys after saving to clear up space on a computer.
        
    # Bad. Some kind of option failure
    return None

def saveGames(steamLibrary, localLibrary, maxBackups, option, uuid):
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
            saveGame(steamLibrary, localLibrary, maxBackups, option, path, uuid)
    else:
        with ThreadPoolExecutor() as executor:
            futures = {
                    executor.submit(saveGame, steamLibrary, localLibrary, maxBackups, option, path, uuid): path for path in savePaths
                    }

        for future in as_completed(futures):
            path = futures[future]
            try:
                result = future.result()  # This retrieves the result of the call
                if result:
                    print(f"Successfully saved game with path: {path}.")
            except Exception as e:
                print(f"Error occurred while saving game for path: {path}. Exception: {e}")

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Used to save and restore steam game save data')
    parser.add_argument('steamLibraryPath', nargs='?', help='The path to your steam library')
    parser.add_argument('-s', '--save', action='store_true', help='Save Steam game data')
    parser.add_argument('-r', '--restore', action='store_true', help='Restore Steam game data')
    parser.add_argument('-l', '--localLibrary', nargs='?', help='Location to store or restore Steam data from or to')
    parser.add_argument('-b', '--backups', type=int, help='The number of Steam game backups to store during saving')
    parser.add_argument('-u', '--uuid', type=int, help='Needed if the game stores save data in the userdata folder and more than one user has directories in the path')
    parser.add_argument('-d', '--debug', action='store_true', help='Used by developers to debug the program')
    parser.add_argument('-c', '--config', nargs='?', help='Config file to reduce the required args')
    args = parser.parse_args()
    # Example steamLibrary path /media/<user>/<drive>/steamLibrary/
    # Add argParser with required steamLibrary path but other optional

    # Load the configuration file if provided
    # Basic configuration might need updating
    if args.config != None and os.path.exists(args.config):
        with open(args.config) as f:
            config = yaml.safe_load(f)
        
        # Update the command-line arguments with the config values if not provided
        args.steamLibraryPath = args.steamLibraryPath or config.get('steamLibraryPath')
        args.save = args.save or config.get('save', False)
        args.restore = args.restore or config.get('restore', False)
        args.localLibrary = args.localLibrary or config.get('localLibrary')
        args.backups = args.backups if args.backups is not None else config.get('backups')
        args.uuid = args.uuid if args.uuid is not None else config.get('uuid')
        args.debug = args.debug or config.get('debug', False)

    if args.debug:
        debugMode = True
        log(args)

    if not args.steamLibraryPath[-1] == "/":
        args.steamLibraryPath = f"{args.steamLibraryPath}/"
    
    if "~" in args.steamLibraryPath:
        args.steamLibraryPath = os.path.expanduser(args.steamLibraryPath)

    if args.backups == None:
        args.backups = 2 # default
    if args.localLibrary == None:
        args.localLibrary = "../SteamSaveLocal/"
    else:
        if not args.localLibrary[-1] == "/":
            args.localLibrary = f"{args.localLibrary}/"

    if not os.path.exists(args.localLibrary):
        os.mkdir(args.localLibrary)

    if args.uuid == None:
        pathWithUUID = f"{args.steamLibraryPath}userdata"
        if os.path.isdir(pathWithUUID):
            folders = [f for f in os.listdir(pathWithUUID) if os.path.isdir(os.path.join(pathWithUUID, f))]

            # Check if there is exactly one folder with the uuids
            if len(folders) == 1:
                args.uuid = folders[0]  # Set uuid to the name of the folder
            else:
                print("Steam UUID not automatically found. Games that require it will not be tried")

    option = None
    if args.save:
        option = "save"

    if args.restore:
        option = "restore"

    log(args)

    saveGames(args.steamLibraryPath, args.localLibrary, args.backups, option, args.uuid)
