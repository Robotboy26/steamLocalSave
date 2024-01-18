import os

def searchDir(dir, target, depth):
    for root, dirs, files in os.walk(dir):
        if target in dirs:
            print(f"Found {target} at: {os.path.join(root, target)}")

        depth -= 1

        if depth <= 0:
            break

# this will search for game save file locations given a game name and (mabey) also just anything that looks like save data
def findSave(gameName):
    
    # some of the most common location are 


