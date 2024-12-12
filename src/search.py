import os

def searchDir(dir, target, depth):
    # do the os.norm thing
    # also if ~ in dir do exspand home
    for root, dirs, files in os.walk(dir):
        if target in dirs:
            foundPath = os.path.join(root, target)
            print(f"Found {target} at: {foundPath}")
            return foundPath

        depth -= 1

        if depth <= 0:
            return None

# this will search for game save file locations given a game name and (mabey) also just anything that looks like save data
def findSaves():
    
    # linux
    # some of the most common location are 

    for root, dirs, files in os.walk("/media/robot/steamgames/SteamLibrary/compatData"):
        print(dirs)
    # gameSave = searchDir("~/SteamLib/compactData/<id>/a bunch of stuff/AppData/something/GameSaveLocation (Hopefully)", gameName, 2)


if __name__ == "__main__":
    findSaves()


